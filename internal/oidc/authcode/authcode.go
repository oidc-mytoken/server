package authcode

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/httpclient"
	"github.com/oidc-mytoken/utils/unixtime"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/oidc-mytoken/utils/utils/jwtutils"
	"github.com/oidc-mytoken/utils/utils/ternary"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/accesstokenrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/oidc/oidcreqres"
	"github.com/oidc-mytoken/server/internal/oidc/pkce"
	provider2 "github.com/oidc-mytoken/server/internal/oidc/provider"
	"github.com/oidc-mytoken/server/internal/server/httpstatus"
	"github.com/oidc-mytoken/server/internal/server/routes"
	iutils "github.com/oidc-mytoken/server/internal/utils"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// GetAuthorizationURL creates an authorization url
func GetAuthorizationURL(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, provider model.Provider, oState *state.State,
	restrictions restrictions.Restrictions,
) (string, error) {
	rlog.Debug("Generating authorization url")
	pkceCode := pkce.NewS256PKCE(utils.RandASCIIString(44))
	if err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return authcodeinforepo.SetCodeVerifier(rlog, tx, oState, pkceCode.Verifier())
		},
	); err != nil {
		return "", err
	}
	pkceChallenge, _ := pkceCode.Challenge()
	return provider.GetAuthorizationURL(
		rlog, oState.State(), pkceChallenge, restrictions.GetScopes(), restrictions.GetAudiences(),
	)
}

func trustedRedirectURI(redirectURI string) bool {
	for _, r := range config.Get().Features.OIDCFlows.AuthCode.Web.TrustedRedirectsRegex {
		if r.MatchString(redirectURI) {
			return true
		}
	}
	return false
}

// StartAuthCodeFlow starts an authorization code flow
func StartAuthCodeFlow(ctx *fiber.Ctx, req *response.AuthCodeFlowRequest) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle authcode")
	native := req.Native() && config.Get().Features.Polling.Enabled
	if !native && req.RedirectURI == "" {
		return &model.Response{
			Status: fiber.StatusBadRequest,
			Response: api.Error{
				Error:            api.ErrorStrInvalidRequest,
				ErrorDescription: "parameter redirect_uri must be given for client_type=web",
			},
		}
	}
	req.Restrictions.ReplaceThisIP(ctx.IP())
	req.Restrictions.ClearUnsupportedKeys()
	p := provider2.GetProvider(req.Issuer)
	if p == nil {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}
	}
	req.Issuer = p.Issuer()
	exp := req.Restrictions.GetExpires()
	if exp > 0 && exp < unixtime.Now() {
		return model.BadRequestErrorResponse("token would already be expired")
	}

	oState, consentCode := state.CreateState()
	authFlowInfo := authcodeinforepo.AuthFlowInfo{
		AuthFlowInfoOut: authcodeinforepo.AuthFlowInfoOut{
			State:               oState,
			AuthCodeFlowRequest: *req,
		},
	}
	res := api.AuthCodeFlowResponse{
		ConsentURI: utils.CombineURLPath(routes.ConsentEndpoint, consentCode.String()),
	}
	if native {
		poll := authFlowInfo.State.PollingCode(rlog)
		authFlowInfo.PollingCode = transfercoderepo.CreatePollingCode(poll, req.ResponseType, req.MaxTokenLen)
		res.PollingInfo = api.PollingInfo{
			PollingCode:          poll,
			PollingCodeExpiresIn: config.Get().Features.Polling.PollingCodeExpiresAfter,
			PollingInterval:      config.Get().Features.Polling.PollingInterval,
		}
	}
	if err := authFlowInfo.Store(rlog, nil); err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if !native && trustedRedirectURI(req.RedirectURI) {
		authURI, err := GetAuthorizationURL(
			rlog, nil, p, state.NewState(consentCode.GetState()),
			req.Restrictions.Restrictions,
		)
		if err != nil {
			rlog.Errorf("%s", errorfmt.Full(err))
			return model.ErrorToInternalServerErrorResponse(err)
		}
		return &model.Response{
			Status: httpstatus.StatusOKForward,
			Response: map[string]string{
				"authorization_uri": authURI,
			},
		}
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: res,
	}
}

// CodeExchange performs an oidc code exchange it creates the mytoken and stores it in the database
func CodeExchange(
	rlog log.Ext1FieldLogger, oState *state.State, code string, networkData api.ClientMetaData,
) *model.Response {
	rlog.Debug("Handle code exchange")
	authInfo, err := authcodeinforepo.GetAuthFlowInfoByState(rlog, oState)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &model.Response{
				Status:   fiber.StatusBadRequest,
				Response: api.ErrorStateMismatch,
			}
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	p := provider2.GetProvider(authInfo.Issuer)
	if p == nil {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}
	}
	params := url.Values{}
	params.Set("grant_type", "authorization_code")
	params.Set("code_verifier", authInfo.CodeVerifier)
	params.Set("code", code)
	params.Set("redirect_uri", routes.RedirectURI)
	params.Set("client_id", p.ClientID())

	httpRes, err := p.AddClientAuthentication(httpclient.Do().R(), p.Endpoints().Token).
		SetFormDataFromValues(params).
		SetResult(&oidcreqres.OIDCTokenResponse{}).
		SetError(&oidcreqres.OIDCErrorResponse{}).
		Post(p.Endpoints().Token)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if errRes, ok := httpRes.Error().(*oidcreqres.OIDCErrorResponse); ok && errRes != nil && errRes.Error != "" {
		return &model.Response{
			Status:   httpRes.RawResponse.StatusCode,
			Response: model.OIDCError(errRes.Error, errRes.ErrorDescription),
		}
	}
	oidcTokenRes, ok := httpRes.Result().(*oidcreqres.OIDCTokenResponse)
	if !ok {
		return &model.Response{
			Status:   httpstatus.StatusOIDPError,
			Response: model.ErrorWithoutDescription("could not unmarshal OP response"),
		}
	}

	if oidcTokenRes.RefreshToken == "" {
		return &model.Response{
			Status:   httpstatus.StatusOIDPError,
			Response: api.ErrorNoRefreshToken,
		}
	}
	scopes := authInfo.Restrictions.GetScopes()
	scopesStr := oidcTokenRes.Scopes
	if scopesStr != "" {
		scopes = iutils.SplitIgnoreEmpty(scopesStr, " ")
		authInfo.Restrictions.SetMaxScopes(scopes) // Update restrictions with correct scopes
	}
	audiences := authInfo.Restrictions.GetAudiences()
	if tmp, ok := jwtutils.GetAudiencesFromJWT(rlog, oidcTokenRes.AccessToken); ok {
		audiences = tmp
	}
	authInfo.Restrictions.SetMaxAudiences(audiences) // Update restrictions with correct audiences

	oidcSub, ok := jwtutils.GetStringFromJWT(rlog, oidcTokenRes.IDToken, "sub")
	if !ok {
		return &model.Response{
			Status:   httpstatus.StatusOIDPError,
			Response: model.ErrorWithoutDescription("could not get 'subject' from id token"),
		}
	}
	var ste *mytokenrepo.MytokenEntry
	if err = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			ste, err = createMytokenEntry(rlog, tx, authInfo, oidcTokenRes.RefreshToken, oidcSub, networkData)
			if err != nil {
				return err
			}
			at := accesstokenrepo.AccessToken{
				Token:     oidcTokenRes.AccessToken,
				IP:        networkData.IP,
				Comment:   "Initial Access Token from authorization code flow",
				Mytoken:   ste.Token,
				Scopes:    scopes,
				Audiences: audiences,
			}
			if err = at.Store(rlog, tx); err != nil {
				return err
			}
			if authInfo.PollingCode {
				jwt, err := ste.Token.ToJWT()
				if err != nil {
					return err
				}
				if err = transfercoderepo.LinkPollingCodeToMT(
					rlog, tx, oState.PollingCode(rlog), jwt, ste.ID,
				); err != nil {
					return err
				}
			}
			return authcodeinforepo.DeleteAuthFlowInfoByState(rlog, tx, oState)
		},
	); err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if authInfo.PollingCode {
		uri := "/native"
		if authInfo.ApplicationName != "" {
			uri = fmt.Sprintf("%s?application=%s", uri, authInfo.ApplicationName)
		}
		return &model.Response{
			Status:   fiber.StatusSeeOther,
			Response: uri,
		}
	}
	res, err := ste.Token.ToTokenResponse(rlog, authInfo.ResponseType, authInfo.MaxTokenLen, networkData, "")
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	var cookie *fiber.Cookie
	if authInfo.ResponseType == model.ResponseTypeTransferCode {
		cookie = cookies.TransferCodeCookie(res.TransferCode, int(res.ExpiresIn))
	} else {
		cookie = cookies.MytokenCookie(res.Mytoken)
	}
	return &model.Response{
		Status:   fiber.StatusSeeOther,
		Response: ternary.IfNotEmptyOr(authInfo.RedirectURI, "/home"),
		Cookies:  []*fiber.Cookie{cookie},
	}
}

func createMytokenEntry(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, authFlowInfo *authcodeinforepo.AuthFlowInfoOut, rt,
	oidcSub string, networkData api.ClientMetaData,
) (*mytokenrepo.MytokenEntry, error) {
	var rot *api.Rotation
	if authFlowInfo.Rotation != nil {
		rot = &authFlowInfo.Rotation.Rotation
	}
	mt, err := mytoken.NewMytoken(
		oidcSub,
		authFlowInfo.Issuer,
		authFlowInfo.Name,
		authFlowInfo.Restrictions.Restrictions,
		authFlowInfo.Capabilities.Capabilities,
		rot,
		unixtime.Now(),
	)
	if err != nil {
		return nil, err
	}
	mte := mytokenrepo.NewMytokenEntry(mt, authFlowInfo.Name, networkData)
	mte.Token.AuthTime = unixtime.Now()
	if err = mte.InitRefreshToken(rt); err != nil {
		return nil, err
	}
	if err = mte.Store(rlog, tx, "Used grant_type oidc_flow authorization_code"); err != nil {
		return nil, err
	}
	return mte, nil
}
