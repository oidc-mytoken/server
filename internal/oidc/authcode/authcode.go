package authcode

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/context"
	"github.com/oidc-mytoken/utils/unixtime"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/oidc-mytoken/utils/utils/issuerutils"
	"github.com/oidc-mytoken/utils/utils/jwtutils"
	"github.com/oidc-mytoken/utils/utils/ternary"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

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
	"github.com/oidc-mytoken/server/internal/oidc/issuer"
	"github.com/oidc-mytoken/server/internal/oidc/pkce"
	"github.com/oidc-mytoken/server/internal/server/httpstatus"
	"github.com/oidc-mytoken/server/internal/server/routes"
	iutils "github.com/oidc-mytoken/server/internal/utils"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

var redirectURI string
var consentEndpoint string

// Init initializes the authcode component
func Init() {
	generalPaths := routes.GetGeneralPaths()
	redirectURI = utils.CombineURLPath(config.Get().IssuerURL, generalPaths.OIDCRedirectEndpoint)
	consentEndpoint = utils.CombineURLPath(config.Get().IssuerURL, generalPaths.ConsentEndpoint)
}

// GetAuthorizationURL creates a authorization url
func GetAuthorizationURL(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, provider *config.ProviderConf, oState *state.State,
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
	scopes := restrictions.GetScopes()
	if len(scopes) <= 0 {
		scopes = provider.Scopes
	}
	oauth2Config := oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		Endpoint:     provider.Endpoints.OAuth2(),
		RedirectURL:  redirectURI,
		Scopes:       scopes,
	}
	additionalParams := []oauth2.AuthCodeOption{
		oauth2.ApprovalForce,
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", pkce.TransformationS256.String()),
	}
	if issuerutils.CompareIssuerURLs(provider.Issuer, issuer.GOOGLE) {
		additionalParams = append(additionalParams, oauth2.AccessTypeOffline)
	} else if !utils.StringInSlice(oidc.ScopeOfflineAccess, oauth2Config.Scopes) {
		oauth2Config.Scopes = append(oauth2Config.Scopes, oidc.ScopeOfflineAccess)
	}
	// Even if user deselected openid scope in restriction, we still need it
	if !utils.StringInSlice(oidc.ScopeOpenID, oauth2Config.Scopes) {
		oauth2Config.Scopes = append(oauth2Config.Scopes, oidc.ScopeOpenID)
	}
	auds := restrictions.GetAudiences()
	if len(auds) > 0 {
		if provider.Audience.SpaceSeparateAuds {
			additionalParams = append(
				additionalParams,
				oauth2.SetAuthURLParam(provider.Audience.RequestParameter, strings.Join(auds, " ")),
			)
		} else {
			for _, a := range auds {
				additionalParams = append(
					additionalParams, oauth2.SetAuthURLParam(provider.Audience.RequestParameter, a),
				)
			}
		}
	}

	return oauth2Config.AuthCodeURL(oState.State(), additionalParams...), nil
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
	provider, ok := config.Get().ProviderByIssuer[req.Issuer]
	if !ok {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}
	}
	req.Issuer = provider.Issuer
	exp := req.Restrictions.GetExpires()
	if exp > 0 && exp < unixtime.Now() {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError("token would already be expired"),
		}
	}

	oState, consentCode := state.CreateState()
	authFlowInfo := authcodeinforepo.AuthFlowInfo{
		AuthFlowInfoOut: authcodeinforepo.AuthFlowInfoOut{
			State:               oState,
			AuthCodeFlowRequest: *req,
		},
	}
	res := api.AuthCodeFlowResponse{
		ConsentURI: utils.CombineURLPath(consentEndpoint, consentCode.String()),
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
			rlog, nil, provider, state.NewState(consentCode.GetState()),
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
	provider, ok := config.Get().ProviderByIssuer[authInfo.Issuer]
	if !ok {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}
	}
	oauth2Config := oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		Endpoint:     provider.Endpoints.OAuth2(),
		RedirectURL:  redirectURI,
	}
	token, err := oauth2Config.Exchange(
		context.Get(), code, oauth2.SetAuthURLParam("code_verifier", authInfo.CodeVerifier),
	)
	if err != nil {
		var e *oauth2.RetrieveError
		if errors.As(err, &e) {
			res, resOK := model.OIDCErrorFromBody(e.Body)
			if !resOK {
				res = model.OIDCError(e.Error(), "")
			}
			rlog.WithError(e).Error("error in code exchange")
			return &model.Response{
				Status:   e.Response.StatusCode,
				Response: res,
			}
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if token.RefreshToken == "" {
		return &model.Response{
			Status:   fiber.StatusInternalServerError,
			Response: api.ErrorNoRefreshToken,
		}
	}
	scopes := authInfo.Restrictions.GetScopes()
	scopesStr, ok := token.Extra("scope").(string)
	if ok && scopesStr != "" {
		scopes = iutils.SplitIgnoreEmpty(scopesStr, " ")
		authInfo.Restrictions.SetMaxScopes(scopes) // Update restrictions with correct scopes
	}
	audiences := authInfo.Restrictions.GetAudiences()
	if tmp, ok := jwtutils.GetAudiencesFromJWT(rlog, token.AccessToken); ok {
		audiences = tmp
	}
	authInfo.Restrictions.SetMaxAudiences(audiences) // Update restrictions with correct audiences

	oidcSub, err := getSubjectFromUserinfo(provider.Provider, token)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	var ste *mytokenrepo.MytokenEntry
	if err = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			ste, err = createMytokenEntry(rlog, tx, authInfo, token, oidcSub, networkData)
			if err != nil {
				return err
			}
			at := accesstokenrepo.AccessToken{
				Token:     token.AccessToken,
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
		url := "/native"
		if authInfo.ApplicationName != "" {
			url = fmt.Sprintf("%s?application=%s", url, authInfo.ApplicationName)
		}
		return &model.Response{
			Status:   fiber.StatusSeeOther,
			Response: url,
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
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, authFlowInfo *authcodeinforepo.AuthFlowInfoOut, token *oauth2.Token,
	oidcSub string, networkData api.ClientMetaData,
) (*mytokenrepo.MytokenEntry, error) {
	var rot *api.Rotation
	if authFlowInfo.Rotation != nil {
		rot = &authFlowInfo.Rotation.Rotation
	}
	mte := mytokenrepo.NewMytokenEntry(
		mytoken.NewMytoken(
			oidcSub,
			authFlowInfo.Issuer,
			authFlowInfo.Name,
			authFlowInfo.Restrictions.Restrictions,
			authFlowInfo.Capabilities.Capabilities,
			rot,
			unixtime.Now(),
		),
		authFlowInfo.Name, networkData,
	)
	mte.Token.AuthTime = unixtime.Now()
	if err := mte.InitRefreshToken(token.RefreshToken); err != nil {
		return nil, err
	}
	if err := mte.Store(rlog, tx, "Used grant_type oidc_flow authorization_code"); err != nil {
		return nil, err
	}
	return mte, nil
}

func getSubjectFromUserinfo(provider *oidc.Provider, token *oauth2.Token) (string, error) {
	userInfo, err := provider.UserInfo(context.Get(), oauth2.StaticTokenSource(token))
	if err != nil {
		return "", errors.Wrap(err, "failed to get userinfo")
	}
	return userInfo.Subject, nil
}
