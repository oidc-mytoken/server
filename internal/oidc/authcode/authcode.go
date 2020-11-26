package authcode

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/db/dbModels"
	response "github.com/zachmann/mytoken/internal/endpoints/token/super/pkg"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/oidc/issuer"
	"github.com/zachmann/mytoken/internal/supertoken/restrictions"
	"github.com/zachmann/mytoken/internal/utils"
	"github.com/zachmann/mytoken/internal/utils/issuerUtils"
	"github.com/zachmann/mytoken/internal/utils/oidcUtils"
)

var redirectURL string

func Init() {
	redirectURL = utils.CombineURLPath(config.Get().IssuerURL, "/redirect")
}

const stateLen = 16
const pollingCodeLen = 32

type stateInfo struct {
	Native bool
}

const stateFmt = "%d%s"

func createState(info stateInfo) string {
	r := utils.RandASCIIString(stateLen)
	native := 0
	if info.Native {
		native = 1
	}
	return fmt.Sprintf(stateFmt, native, r)
}

func parseState(state string) stateInfo {
	info := stateInfo{}
	native := 0
	var r string
	fmt.Sscanf(state, stateFmt, &native, &r)
	if native != 0 {
		info.Native = true
	}
	return info
}

func authorizationURL(provider *config.ProviderConf, restrictions restrictions.Restrictions, native bool) (string, string) {
	log.Debug("Generating authorization url")
	scopes := restrictions.GetScopes()
	if len(scopes) <= 0 {
		scopes = provider.Scopes
	}
	oauth2Config := oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		Endpoint:     provider.Provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       scopes,
	}
	state := createState(stateInfo{Native: native})
	additionalParams := []oauth2.AuthCodeOption{oauth2.ApprovalForce}
	if issuerUtils.CompareIssuerURLs(provider.Issuer, issuer.GOOGLE) {
		additionalParams = append(additionalParams, oauth2.AccessTypeOffline)
	} else if !utils.StringInSlice(oidc.ScopeOfflineAccess, oauth2Config.Scopes) {
		oauth2Config.Scopes = append(oauth2Config.Scopes, oidc.ScopeOfflineAccess)
	}
	auds := restrictions.GetAudiences()
	if len(auds) > 0 {
		additionalParams = append(additionalParams, oauth2.SetAuthURLParam("audience", strings.Join(auds, " ")))
	}

	return oauth2Config.AuthCodeURL(state, additionalParams...), state
}

func InitAuthCodeFlow(body []byte) *model.Response {
	log.Debug("Handle authcode")
	req := response.NewAuthCodeFlowRequest()
	if err := json.Unmarshal(body, &req); err != nil {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError(err.Error()),
		}
	}
	provider, ok := config.Get().ProviderByIssuer[req.Issuer]
	if !ok {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.APIErrorUnknownIssuer,
		}
	}
	exp := req.Restrictions.GetExpires()
	if exp > 0 && exp < time.Now().Unix() {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError("token would already be expired"),
		}
	}

	authURL, state := authorizationURL(provider, req.Restrictions, req.Native())
	authFlowInfo := dbModels.AuthFlowInfo{
		State:                state,
		Issuer:               provider.Issuer,
		Restrictions:         req.Restrictions,
		Capabilities:         req.Capabilities,
		SubtokenCapabilities: req.SubtokenCapabilities,
		Name:                 req.Name,
	}
	res := response.AuthCodeFlowResponse{
		AuthorizationURL: authURL,
	}
	if req.Native() {
		authFlowInfo.PollingCode = utils.RandASCIIString(pollingCodeLen)
		res.PollingCode = authFlowInfo.PollingCode
		res.PollingCodeExpiresIn = config.Get().Polling.PollingCodeExpiresAfter
		res.PollingInterval = config.Get().Polling.PollingInterval
	}
	if err := authFlowInfo.Store(); err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: res,
	}
}

func CodeExchange(state, code string, networkData model.NetworkData) *model.Response {
	log.Debug("Handle code exchange")
	authInfo, err := dbModels.GetAuthCodeInfoByState(state)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &model.Response{
				Status:   fiber.StatusBadRequest,
				Response: model.APIErrorStateMismatch,
			}
		}
		return model.ErrorToInternalServerErrorResponse(err)
	}
	provider, ok := config.Get().ProviderByIssuer[authInfo.Issuer]
	if !ok {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.APIErrorUnknownIssuer,
		}
	}
	oauth2Config := oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		Endpoint:     provider.Provider.Endpoint(),
		RedirectURL:  redirectURL,
	}
	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		var e *oauth2.RetrieveError
		if errors.As(err, &e) {
			res, resOK := model.OIDCErrorFromBody(e.Body)
			if !resOK {
				res = model.OIDCError(e.Error(), "")
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
			Response: model.APIErrorNoRefreshToken,
		}
	}
	scopes := authInfo.Restrictions.GetScopes()
	scopesStr, ok := token.Extra("scope").(string)
	if ok && scopesStr != "" {
		scopes = utils.SplitIgnoreEmpty(scopesStr, " ")
		authInfo.Restrictions.SetMaxScopes(scopes) // Update restrictions with correct scopes
	}
	audiences := authInfo.Restrictions.GetAudiences()
	if tmp, ok := oidcUtils.GetAudiencesFromJWT(token.AccessToken); ok {
		audiences = tmp
	}
	authInfo.Restrictions.SetMaxAudiences(audiences) // Update restrictions with correct audiences

	oidcSub, err := getSubjectFromUserinfo(provider.Provider, token)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	ste, err := createSuperTokenEntry(authInfo, token, oidcSub, networkData)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	at := dbModels.AccessToken{
		Token:     token.AccessToken,
		IP:        networkData.IP,
		Comment:   "Initial Access Token from authorization code flow",
		STID:      ste.ID,
		Scopes:    scopes,
		Audiences: audiences,
	}
	if err := at.Store(); err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if err := db.Transact(func(tx *sqlx.Tx) error {
		if authInfo.PollingCode != "" {
			if _, err := tx.Exec(`INSERT INTO TmpST (polling_code_id, ST_id) VALUES((SELECT id FROM PollingCodes WHERE polling_code = ?), ?)`, authInfo.PollingCode, ste.ID); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(`DELETE FROM AuthInfo WHERE state = ?`, authInfo.State); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	//TODO on the response the idea was to redirect to a correct side, that has the response
	if authInfo.PollingCode != "" {
		return &model.Response{
			Status:   fiber.StatusOK,
			Response: "ok", //TODO
		}
	}
	res := ste.Token.ToSuperTokenResponse("")
	return &model.Response{
		Status:   fiber.StatusSeeOther,
		Response: "/", //TODO redirect
		Cookies: []*fiber.Cookie{{
			Name:     "mytoken-supertoken",
			Value:    res.SuperToken,
			Path:     "/api",
			MaxAge:   3600,  //TODO from config
			Secure:   false, //TODO true
			HTTPOnly: true,
			SameSite: "Strict",
		}},
	}
}

func createSuperTokenEntry(authFlowInfo *dbModels.AuthFlowInfo, token *oauth2.Token, oidcSub string, networkData model.NetworkData) (*dbModels.SuperTokenEntry, error) {
	ste := dbModels.NewSuperTokenEntry(authFlowInfo.Name, oidcSub, authFlowInfo.Issuer, authFlowInfo.Restrictions, authFlowInfo.Capabilities, authFlowInfo.SubtokenCapabilities, networkData)
	ste.RefreshToken = token.RefreshToken
	err := ste.Store("Used grant_type oidc_flow authorization_code")
	if err != nil {
		return nil, err
	}
	return ste, nil
}

func getSubjectFromUserinfo(provider *oidc.Provider, token *oauth2.Token) (string, error) {
	userInfo, err := provider.UserInfo(context.Background(), oauth2.StaticTokenSource(token))
	if err != nil {
		return "", fmt.Errorf("failed to get userinfo: %s", err)
	}
	return userInfo.Subject, nil
}
