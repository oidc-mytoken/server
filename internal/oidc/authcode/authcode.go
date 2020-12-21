package authcode

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/context"
	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/db/dbrepo/accesstokenrepo"
	"github.com/zachmann/mytoken/internal/db/dbrepo/authcodeinforepo"
	"github.com/zachmann/mytoken/internal/db/dbrepo/pollingcoderepo"
	"github.com/zachmann/mytoken/internal/db/dbrepo/supertokenrepo"
	response "github.com/zachmann/mytoken/internal/endpoints/token/super/pkg"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/oidc/issuer"
	"github.com/zachmann/mytoken/internal/server/routes"
	supertoken "github.com/zachmann/mytoken/internal/supertoken/pkg"
	"github.com/zachmann/mytoken/internal/supertoken/restrictions"
	"github.com/zachmann/mytoken/internal/utils"
	"github.com/zachmann/mytoken/internal/utils/issuerUtils"
	"github.com/zachmann/mytoken/internal/utils/oidcUtils"
)

var redirectURL string

// Init initializes the authcode component
func Init() {
	redirectURL = utils.CombineURLPath(config.Get().IssuerURL, routes.GetGeneralPaths().OIDCRedirectEndpoint)
}

const stateLen = 16
const pollingCodeLen = 32

type stateInfo struct {
	Native       bool
	ResponseType model.ResponseType
}

type pollingInfo struct {
	ResponseType model.ResponseType
}

const stateFmt = "%d:%d:%s"
const pollingFmt = "%d:%s"

func createPollingCode(info pollingInfo) string {
	r := utils.RandASCIIString(pollingCodeLen)
	return fmt.Sprintf(pollingFmt, info.ResponseType, r)
}

// ParsePollingCode parses a polling code string into pollingInfo
func ParsePollingCode(pollingCode string) pollingInfo {
	info := pollingInfo{}
	var r string
	fmt.Sscanf(pollingCode, pollingFmt, &info.ResponseType, &r)
	return info
}

func createState(info stateInfo) string {
	r := utils.RandASCIIString(stateLen)
	native := 0
	if info.Native {
		native = 1
	}
	return fmt.Sprintf(stateFmt, native, info.ResponseType, r)
}

func parseState(state string) stateInfo {
	info := stateInfo{}
	native := 0
	var r string
	fmt.Sscanf(state, stateFmt, &native, &info.ResponseType, &r)
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
		Endpoint:     provider.Endpoints.OAuth2(),
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

// StartAuthCodeFlow starts an authorization code flow
func StartAuthCodeFlow(body []byte) *model.Response {
	log.Debug("Handle authcode")
	req := response.NewAuthCodeFlowRequest()
	if err := json.Unmarshal(body, &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
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
	authFlowInfo := authcodeinforepo.AuthFlowInfo{
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
	if req.Native() && config.Get().Features.Polling.Enabled {
		authFlowInfo.PollingCode = createPollingCode(pollingInfo{ResponseType: req.ResponseType})
		res.PollingCode = authFlowInfo.PollingCode
		res.PollingCodeExpiresIn = config.Get().Features.Polling.PollingCodeExpiresAfter
		res.PollingInterval = config.Get().Features.Polling.PollingInterval
	}
	if err := authFlowInfo.Store(nil); err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: res,
	}
}

// CodeExchange performs an oidc code exchange it creates the super token and stores it in the database
func CodeExchange(state, code string, networkData model.ClientMetaData) *model.Response {
	log.Debug("Handle code exchange")
	authInfo, err := authcodeinforepo.GetAuthFlowInfoByState(state)
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
		Endpoint:     provider.Endpoints.OAuth2(),
		RedirectURL:  redirectURL,
	}
	token, err := oauth2Config.Exchange(context.Get(), code)
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
	var ste *supertokenrepo.SuperTokenEntry
	if err = db.Transact(func(tx *sqlx.Tx) error {
		ste, err = createSuperTokenEntry(tx, authInfo, token, oidcSub, networkData)
		if err != nil {
			return err
		}
		at := accesstokenrepo.AccessToken{
			Token:     token.AccessToken,
			IP:        networkData.IP,
			Comment:   "Initial Access Token from authorization code flow",
			STID:      ste.ID,
			Scopes:    scopes,
			Audiences: audiences,
		}
		if err = at.Store(tx); err != nil {
			return err
		}
		if authInfo.PollingCode != "" {
			if err = pollingcoderepo.LinkPollingCodeToST(tx, authInfo.PollingCode, ste.ID); err != nil {
				return err
			}
		}
		if err = authcodeinforepo.DeleteAuthFlowInfoByState(tx, authInfo.State); err != nil {
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
			Response: "ok", // TODO
		}
	}
	stateInf := parseState(state)
	res, err := ste.Token.ToTokenResponse(stateInf.ResponseType, networkData, "")
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	cookieName := "mytoken-supertoken"
	cookieValue := res.SuperToken
	cookieAge := 3600 //TODO from config
	if stateInf.ResponseType == model.ResponseTypeTransferCode {
		cookieName = "mytoken-transfercode"
		cookieValue = res.TransferCode
		cookieAge = int(res.ExpiresIn)
	}
	return &model.Response{
		Status:   fiber.StatusSeeOther,
		Response: "/", //TODO redirect
		Cookies: []*fiber.Cookie{{
			Name:     cookieName,
			Value:    cookieValue,
			Path:     "/api",
			MaxAge:   cookieAge,
			Secure:   false, //TODO depending on TLS
			HTTPOnly: true,
			SameSite: "Strict",
		}},
	}
}

func createSuperTokenEntry(tx *sqlx.Tx, authFlowInfo *authcodeinforepo.AuthFlowInfo, token *oauth2.Token, oidcSub string, networkData model.ClientMetaData) (*supertokenrepo.SuperTokenEntry, error) {
	ste := supertokenrepo.NewSuperTokenEntry(
		supertoken.NewSuperToken(
			oidcSub,
			authFlowInfo.Issuer,
			authFlowInfo.Restrictions,
			authFlowInfo.Capabilities,
			authFlowInfo.SubtokenCapabilities),
		authFlowInfo.Name, networkData)
	ste.RefreshToken = token.RefreshToken
	err := ste.Store(tx, "Used grant_type oidc_flow authorization_code")
	if err != nil {
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
