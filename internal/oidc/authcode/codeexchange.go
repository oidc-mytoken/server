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
	"github.com/oidc-mytoken/utils/utils/jwtutils"
	"github.com/oidc-mytoken/utils/utils/ternary"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/accesstokenrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/userrepo"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo"
	"github.com/oidc-mytoken/server/internal/db/profilerepo"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/oidc/oidcreqres"
	provider2 "github.com/oidc-mytoken/server/internal/oidc/provider"
	"github.com/oidc-mytoken/server/internal/oidc/userinfo"
	"github.com/oidc-mytoken/server/internal/server/httpstatus"
	"github.com/oidc-mytoken/server/internal/server/routes"
	iutils "github.com/oidc-mytoken/server/internal/utils"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

// CodeExchange performs an OIDC code exchange, creates the mytoken, and stores it in the database.
func CodeExchange(
	rlog log.Ext1FieldLogger, oState *state.State, code string, networkData api.ClientMetaData,
) *model.Response {
	rlog.Debug("Handle code exchange")
	authInfo, errRes := fetchAuthInfo(rlog, oState)
	if errRes != nil {
		return errRes
	}
	p, errRes := fetchProvider(authInfo.Issuer)
	if errRes != nil {
		return errRes
	}
	oidcTokenRes, errRes := exchangeCodeForToken(rlog, p, authInfo, code)
	if errRes != nil {
		return errRes
	}

	updateAuthInfoScopesAndAudiences(rlog, authInfo, oidcTokenRes)
	userInfos, errRes := fetchUserInfos(rlog, p, oidcTokenRes)
	if errRes != nil {
		return errRes
	}
	enforcedRestrictions, errRes := getEnforcedRestrictionTemplate(
		provider2.GetEnforcedRestrictionsByIssuer(p.Issuer()), userInfos,
	)
	if errRes != nil {
		return errRes
	}
	ste, errRes := storeTokenInDatabase(
		rlog, oState, authInfo, enforcedRestrictions, oidcTokenRes, userInfos, networkData,
	)
	if errRes != nil {
		return errRes
	}

	return generateResponse(rlog, authInfo, ste, networkData)
}

func fetchAuthInfo(rlog log.Ext1FieldLogger, oState *state.State) (*authcodeinforepo.AuthFlowInfoOut, *model.Response) {
	authInfo, err := authcodeinforepo.GetAuthFlowInfoByState(rlog, nil, oState)
	if err == nil {
		return authInfo, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorStateMismatch,
		}
	}
	rlog.Errorf("%s", errorfmt.Full(err))
	return nil, model.ErrorToInternalServerErrorResponse(err)
}

func fetchProvider(issuer string) (model.Provider, *model.Response) {
	p := provider2.GetProvider(issuer)
	if p == nil {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}
	}
	return p, nil
}

func exchangeCodeForToken(
	rlog log.Ext1FieldLogger, p model.Provider, authInfo *authcodeinforepo.AuthFlowInfoOut,
	code string,
) (*oidcreqres.OIDCTokenResponse, *model.Response) {
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
		return nil, model.ErrorToInternalServerErrorResponse(err)
	}

	if errRes, ok := httpRes.Error().(*oidcreqres.OIDCErrorResponse); ok && errRes != nil && errRes.Error != "" {
		return nil, &model.Response{
			Status:   httpRes.RawResponse.StatusCode,
			Response: model.OIDCError(errRes.Error, errRes.ErrorDescription),
		}
	}

	oidcTokenRes, ok := httpRes.Result().(*oidcreqres.OIDCTokenResponse)
	if !ok {
		return nil, &model.Response{
			Status:   httpstatus.StatusOIDPError,
			Response: model.ErrorWithoutDescription("could not unmarshal OP response"),
		}
	}

	if oidcTokenRes.RefreshToken == "" {
		return nil, &model.Response{
			Status:   httpstatus.StatusOIDPError,
			Response: api.ErrorNoRefreshToken,
		}
	}

	return oidcTokenRes, nil
}

func updateAuthInfoScopesAndAudiences(
	rlog log.Ext1FieldLogger, authInfo *authcodeinforepo.AuthFlowInfoOut, oidcTokenRes *oidcreqres.OIDCTokenResponse,
) {
	if scopesStr := oidcTokenRes.Scopes; scopesStr != "" {
		scopes := iutils.SplitIgnoreEmpty(scopesStr, " ")
		authInfo.Restrictions.SetMaxScopes(scopes)
	}

	audiences := authInfo.Restrictions.GetAudiences()
	if tmp, ok := jwtutils.GetAudiencesFromJWT(rlog, oidcTokenRes.AccessToken); ok {
		audiences = tmp
	}
	authInfo.Restrictions.SetMaxAudiences(audiences)
}

func fetchUserInfos(rlog log.Ext1FieldLogger, p model.Provider, oidcTokenRes *oidcreqres.OIDCTokenResponse) (
	map[string]any, *model.Response,
) {
	attrs := []string{
		"sub",
		"email",
		"email_verified",
	}
	enforcedRestrictionsConf := provider2.GetEnforcedRestrictionsByIssuer(p.Issuer())
	if enforcedRestrictionsConf.Enabled {
		attrs = append(attrs, enforcedRestrictionsConf.ClaimName)
	}

	userInfos := userinfo.GetUserAttributes(rlog, oidcTokenRes, p, attrs...)
	oidcSub := iutils.GetStringFromAnyMap(userInfos, "sub")
	if oidcSub == "" {
		return nil, &model.Response{
			Status:   httpstatus.StatusOIDPError,
			Response: model.ErrorWithoutDescription("could not get 'subject' from id token"),
		}
	}

	return userInfos, nil
}

func storeTokenInDatabase(
	rlog log.Ext1FieldLogger, oState *state.State, authInfo *authcodeinforepo.AuthFlowInfoOut,
	enforcedRestrictions string, oidcTokenRes *oidcreqres.OIDCTokenResponse, userInfos map[string]any,
	networkData api.ClientMetaData,
) (*mytokenrepo.MytokenEntry, *model.Response) {
	var ste *mytokenrepo.MytokenEntry
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			var err error
			ste, err = createMytokenEntry(
				rlog, tx, authInfo, enforcedRestrictions, oidcTokenRes.RefreshToken,
				iutils.GetStringFromAnyMap(userInfos, "sub"), networkData,
			)
			if err != nil {
				return err
			}
			if err = storeAccessToken(
				rlog, tx, oidcTokenRes.AccessToken, networkData, ste, authInfo.Restrictions.GetScopes(),
				authInfo.Restrictions.GetAudiences(),
			); err != nil {
				return err
			}
			if authInfo.PollingCode {
				if err = linkPollingCodeToMytoken(rlog, tx, oState, ste); err != nil {
					return err
				}
			}
			if err = updateUserMailInfo(rlog, tx, ste.ID, userInfos); err != nil {
				return err
			}
			return authcodeinforepo.DeleteAuthFlowInfoByState(rlog, tx, oState)
		},
	)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return nil, model.ErrorToInternalServerErrorResponse(err)
	}
	return ste, nil
}

func storeAccessToken(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, accessToken string, networkData api.ClientMetaData,
	ste *mytokenrepo.MytokenEntry, scopes []string, audiences []string,
) error {
	at := accesstokenrepo.AccessToken{
		Token:     accessToken,
		IP:        networkData.IP,
		Comment:   "Initial Access Token from authorization code flow",
		Mytoken:   ste.Token,
		Scopes:    scopes,
		Audiences: audiences,
	}
	return at.Store(rlog, tx)
}

func linkPollingCodeToMytoken(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, oState *state.State, ste *mytokenrepo.MytokenEntry,
) error {
	jwt, err := ste.Token.ToJWT()
	if err != nil {
		return err
	}
	return transfercoderepo.LinkPollingCodeToMT(rlog, tx, oState.PollingCode(rlog), jwt, ste.ID)
}

func updateUserMailInfo(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mytokenID mtid.MTID, userInfos map[string]any) error {
	mailInfo, err := userrepo.GetMail(rlog, tx, mytokenID)
	if _, err = db.ParseError(err); err != nil {
		return err
	}
	if mailInfo.Mail.Valid {
		return nil
	}
	// no valid mail in db
	mail := iutils.GetStringFromAnyMap(userInfos, "email")
	mailVerified := iutils.GetBoolFromAnyMap(userInfos, "email_verified")
	return userrepo.SetEmail(rlog, tx, mytokenID, mail, mailVerified)
}

func generateResponse(
	rlog log.Ext1FieldLogger, authInfo *authcodeinforepo.AuthFlowInfoOut, ste *mytokenrepo.MytokenEntry,
	networkData api.ClientMetaData,
) *model.Response {
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
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, authFlowInfo *authcodeinforepo.AuthFlowInfoOut,
	enforcedRestrictionsTemplate, rt, oidcSub string, networkData api.ClientMetaData,
) (*mytokenrepo.MytokenEntry, error) {
	var rot *api.Rotation
	if authFlowInfo.Rotation != nil {
		rot = &authFlowInfo.Rotation.Rotation
	}
	restr := authFlowInfo.Restrictions.Restrictions
	if enforcedRestrictionsTemplate != "" {
		parser := profilerepo.NewDBProfileParser(rlog)
		enforced, err := parser.ParseRestrictionsTemplate([]byte(enforcedRestrictionsTemplate))
		if err != nil {
			return nil, err
		}
		restr, _ = restrictions.Tighten(rlog, restrictions.NewRestrictionsFromAPI(enforced), restr)
	}
	mt, err := mytoken.NewMytoken(
		oidcSub,
		authFlowInfo.Issuer,
		authFlowInfo.Name,
		restr,
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
	if err = notificationsrepo.ScheduleExpirationNotificationsIfNeeded(
		rlog, tx, mte.ID, mte.Token.ExpiresAt, mte.Token.IssuedAt,
	); err != nil {
		return nil, err
	}
	return mte, nil
}
