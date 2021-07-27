package authcode

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/accesstokenrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/issuer"
	"github.com/oidc-mytoken/server/internal/server/routes"
	"github.com/oidc-mytoken/server/shared/context"
	pkgModel "github.com/oidc-mytoken/server/shared/model"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
	"github.com/oidc-mytoken/server/shared/utils"
	"github.com/oidc-mytoken/server/shared/utils/issuerUtils"
	"github.com/oidc-mytoken/server/shared/utils/jwtutils"
	"github.com/oidc-mytoken/server/shared/utils/unixtime"
)

var redirectURL string
var consentEndpoint string

// Init initializes the authcode component
func Init() {
	generalPaths := routes.GetGeneralPaths()
	redirectURL = utils.CombineURLPath(config.Get().IssuerURL, generalPaths.OIDCRedirectEndpoint)
	consentEndpoint = utils.CombineURLPath(config.Get().IssuerURL, generalPaths.ConsentEndpoint)
}

// GetAuthorizationURL creates a authorization url
func GetAuthorizationURL(provider *config.ProviderConf, oState string, restrictions restrictions.Restrictions) string {
	log.Debug("Generating authorization url")
	scopes := restrictions.GetScopes()
	if len(scopes) <= 0 {
		scopes = provider.Scopes
	}
	oauth2Config := oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		Endpoint:     provider.Endpoints.OAuth2(),
		RedirectURL:  redirectURL,
		Scopes:       scopes,
	}
	additionalParams := []oauth2.AuthCodeOption{oauth2.ApprovalForce}
	if issuerUtils.CompareIssuerURLs(provider.Issuer, issuer.GOOGLE) {
		additionalParams = append(additionalParams, oauth2.AccessTypeOffline)
	} else if !utils.StringInSlice(oidc.ScopeOfflineAccess, oauth2Config.Scopes) {
		oauth2Config.Scopes = append(oauth2Config.Scopes, oidc.ScopeOfflineAccess)
	}
	auds := restrictions.GetAudiences()
	if len(auds) > 0 {
		additionalParams = append(additionalParams, oauth2.SetAuthURLParam(provider.AudienceRequestParameter, strings.Join(auds, " ")))
	}

	return oauth2Config.AuthCodeURL(oState, additionalParams...)
}

// StartAuthCodeFlow starts an authorization code flow
func StartAuthCodeFlow(ctx *fiber.Ctx, oidcReq response.OIDCFlowRequest) *model.Response {
	log.Debug("Handle authcode")
	req := oidcReq.ToAuthCodeFlowRequest()
	req.Restrictions.ReplaceThisIp(ctx.IP())
	req.Restrictions.ClearUnsupportedKeys()
	provider, ok := config.Get().ProviderByIssuer[req.Issuer]
	if !ok {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}
	}
	exp := req.Restrictions.GetExpires()
	if exp > 0 && exp < unixtime.Now() {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: pkgModel.BadRequestError("token would already be expired"),
		}
	}

	oState, consentCode := state.CreateState()
	authFlowInfoO := authcodeinforepo.AuthFlowInfoOut{
		State:                oState,
		Issuer:               provider.Issuer,
		Restrictions:         req.Restrictions,
		Capabilities:         req.Capabilities,
		SubtokenCapabilities: req.SubtokenCapabilities,
		Name:                 req.Name,
		Rotation:             req.Rotation,
		ResponseType:         req.ResponseType,
		MaxTokenLen:          req.MaxTokenLen,
	}
	authFlowInfo := authcodeinforepo.AuthFlowInfo{
		AuthFlowInfoOut: authFlowInfoO,
	}
	res := api.AuthCodeFlowResponse{
		AuthorizationURL: utils.CombineURLPath(consentEndpoint, consentCode.String()),
	}
	if req.Native() && config.Get().Features.Polling.Enabled {
		poll := authFlowInfo.State.PollingCode()
		authFlowInfo.PollingCode = transfercoderepo.CreatePollingCode(poll, req.ResponseType, req.MaxTokenLen)
		res.PollingInfo = api.PollingInfo{
			PollingCode:          poll,
			PollingCodeExpiresIn: config.Get().Features.Polling.PollingCodeExpiresAfter,
			PollingInterval:      config.Get().Features.Polling.PollingInterval,
		}
	}
	if err := authFlowInfo.Store(nil); err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: res,
	}
}

// CodeExchange performs an oidc code exchange it creates the mytoken and stores it in the database
func CodeExchange(oState *state.State, code string, networkData api.ClientMetaData) *model.Response {
	log.Debug("Handle code exchange")
	authInfo, err := authcodeinforepo.GetAuthFlowInfoByState(oState)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &model.Response{
				Status:   fiber.StatusBadRequest,
				Response: api.ErrorStateMismatch,
			}
		}
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
		RedirectURL:  redirectURL,
	}
	token, err := oauth2Config.Exchange(context.Get(), code)
	if err != nil {
		var e *oauth2.RetrieveError
		if errors.As(err, &e) {
			res, resOK := pkgModel.OIDCErrorFromBody(e.Body)
			if !resOK {
				res = pkgModel.OIDCError(e.Error(), "")
			}
			return &model.Response{
				Status:   e.Response.StatusCode,
				Response: res,
			}
		}
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
		scopes = utils.SplitIgnoreEmpty(scopesStr, " ")
		authInfo.Restrictions.SetMaxScopes(scopes) // Update restrictions with correct scopes
	}
	audiences := authInfo.Restrictions.GetAudiences()
	if tmp, ok := jwtutils.GetAudiencesFromJWT(token.AccessToken); ok {
		audiences = tmp
	}
	authInfo.Restrictions.SetMaxAudiences(audiences) // Update restrictions with correct audiences

	oidcSub, err := getSubjectFromUserinfo(provider.Provider, token)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	var ste *mytokenrepo.MytokenEntry
	if err = db.Transact(func(tx *sqlx.Tx) error {
		ste, err = createMytokenEntry(tx, authInfo, token, oidcSub, networkData)
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
		if err = at.Store(tx); err != nil {
			return err
		}
		if authInfo.PollingCode {
			jwt, err := ste.Token.ToJWT()
			if err != nil {
				return err
			}
			if err = transfercoderepo.LinkPollingCodeToMT(tx, oState.PollingCode(), jwt, ste.ID); err != nil {
				return err
			}
		}
		return authcodeinforepo.DeleteAuthFlowInfoByState(tx, oState)
	}); err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if authInfo.PollingCode {
		return &model.Response{
			Status:   fiber.StatusSeeOther,
			Response: "/native",
		}
	}
	res, err := ste.Token.ToTokenResponse(authInfo.ResponseType, authInfo.MaxTokenLen, networkData, "")
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	var cookie fiber.Cookie
	if authInfo.ResponseType == pkgModel.ResponseTypeTransferCode {
		cookie = cookies.TransferCodeCookie(res.TransferCode, int(res.ExpiresIn))
	} else {
		cookie = cookies.MytokenCookie(res.Mytoken)
	}
	return &model.Response{
		Status:   fiber.StatusSeeOther,
		Response: "/home",
		Cookies:  []*fiber.Cookie{&cookie},
	}
}

func createMytokenEntry(tx *sqlx.Tx, authFlowInfo *authcodeinforepo.AuthFlowInfoOut, token *oauth2.Token, oidcSub string, networkData api.ClientMetaData) (*mytokenrepo.MytokenEntry, error) {
	ste := mytokenrepo.NewMytokenEntry(
		mytoken.NewMytoken(
			oidcSub,
			authFlowInfo.Issuer,
			authFlowInfo.Restrictions,
			authFlowInfo.Capabilities,
			authFlowInfo.SubtokenCapabilities,
			authFlowInfo.Rotation),
		authFlowInfo.Name, networkData)
	if err := ste.InitRefreshToken(token.RefreshToken); err != nil {
		return nil, err
	}
	if err := ste.Store(tx, "Used grant_type oidc_flow authorization_code"); err != nil {
		return nil, err
	}
	return ste, nil
}

func getSubjectFromUserinfo(provider *oidc.Provider, token *oauth2.Token) (string, error) {
	userInfo, err := provider.UserInfo(context.Get(), oauth2.StaticTokenSource(token))
	if err != nil {
		return "", fmt.Errorf("failed to get userinfo: %s", err)
	}
	return userInfo.Subject, nil
}
