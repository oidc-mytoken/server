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

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/accesstokenrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/supertokenrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/supertokenrepo/transfercoderepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/super/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/issuer"
	"github.com/oidc-mytoken/server/internal/server/routes"
	supertoken "github.com/oidc-mytoken/server/internal/supertoken/pkg"
	"github.com/oidc-mytoken/server/internal/supertoken/restrictions"
	"github.com/oidc-mytoken/server/internal/utils/issuerUtils"
	"github.com/oidc-mytoken/server/internal/utils/oidcUtils"
	"github.com/oidc-mytoken/server/internal/utils/singleasciiencode"
	pkgModel "github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/context"
	"github.com/oidc-mytoken/server/shared/utils"
)

var redirectURL string
var nativeRedirectEndpoint string

// Init initializes the authcode component
func Init() {
	generalPaths := routes.GetGeneralPaths()
	redirectURL = utils.CombineURLPath(config.Get().IssuerURL, generalPaths.OIDCRedirectEndpoint)
	nativeRedirectEndpoint = utils.CombineURLPath(config.Get().IssuerURL, generalPaths.NativeRedirectEndpoint)
}

const stateLen = 16
const pollingCodeLen = 32

type stateInfo struct {
	Native       bool
	ResponseType pkgModel.ResponseType
}

const stateFmt = "%d:%d:%s"

// StateInfoFlags
const (
	flagNative = 0x01
)

func createState(info stateInfo) *state.State {
	r := utils.RandASCIIString(stateLen)
	fe := singleasciiencode.NewFlagEncoder()
	fe.Set("native", info.Native)
	flags := fe.Encode()
	responseType := singleasciiencode.EncodeNumber64(byte(info.ResponseType))
	s := string(append([]byte(r), flags, responseType))
	return state.NewState(s)
}

func parseState(state *state.State) stateInfo {
	info := stateInfo{}
	responseType, _ := singleasciiencode.DecodeNumber64(state.State()[len(state.State())-1])
	flags := singleasciiencode.Decode(state.State()[len(state.State())-2], "native")
	info.ResponseType = pkgModel.ResponseType(responseType)
	info.Native, _ = flags.Get("native")
	return info
}

func authorizationURL(provider *config.ProviderConf, restrictions restrictions.Restrictions, native bool) (string, *state.State) {
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

	return oauth2Config.AuthCodeURL(state.State(), additionalParams...), state
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
			Response: pkgModel.APIErrorUnknownIssuer,
		}
	}
	exp := req.Restrictions.GetExpires()
	if exp > 0 && exp < time.Now().Unix() {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: pkgModel.BadRequestError("token would already be expired"),
		}
	}

	authURL, state := authorizationURL(provider, req.Restrictions, req.Native())
	authFlowInfoO := authcodeinforepo.AuthFlowInfoOut{
		State:                state,
		Issuer:               provider.Issuer,
		Restrictions:         req.Restrictions,
		Capabilities:         req.Capabilities,
		SubtokenCapabilities: req.SubtokenCapabilities,
		Name:                 req.Name,
	}
	authFlowInfo := authcodeinforepo.AuthFlowInfo{
		AuthFlowInfoOut: authFlowInfoO,
	}
	res := response.AuthCodeFlowResponse{
		AuthorizationURL: authURL,
	}
	if req.Native() && config.Get().Features.Polling.Enabled {
		poll := authFlowInfo.State.PollingCode()
		authFlowInfo.PollingCode = transfercoderepo.CreatePollingCode(poll, req.ResponseType, authURL)
		res = response.AuthCodeFlowResponse{
			AuthorizationURL: utils.CombineURLPath(nativeRedirectEndpoint, poll),
			PollingInfo: response.PollingInfo{
				PollingCode:          poll,
				PollingCodeExpiresIn: config.Get().Features.Polling.PollingCodeExpiresAfter,
				PollingInterval:      config.Get().Features.Polling.PollingInterval,
			},
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

// CodeExchange performs an oidc code exchange it creates the super token and stores it in the database
func CodeExchange(state *state.State, code string, networkData model.ClientMetaData) *model.Response {
	log.Debug("Handle code exchange")
	authInfo, err := authcodeinforepo.GetAuthFlowInfoByState(state)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &model.Response{
				Status:   fiber.StatusBadRequest,
				Response: pkgModel.APIErrorStateMismatch,
			}
		}
		return model.ErrorToInternalServerErrorResponse(err)
	}
	provider, ok := config.Get().ProviderByIssuer[authInfo.Issuer]
	if !ok {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: pkgModel.APIErrorUnknownIssuer,
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
			Response: pkgModel.APIErrorNoRefreshToken,
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
			Token:      token.AccessToken,
			IP:         networkData.IP,
			Comment:    "Initial Access Token from authorization code flow",
			SuperToken: ste.Token,
			Scopes:     scopes,
			Audiences:  audiences,
		}
		if err = at.Store(tx); err != nil {
			return err
		}
		if authInfo.PollingCode {
			jwt, err := ste.Token.ToJWT()
			if err != nil {
				return err
			}
			if err = transfercoderepo.LinkPollingCodeToST(tx, state.PollingCode(), jwt); err != nil {
				return err
			}
		}
		if err = authcodeinforepo.DeleteAuthFlowInfoByState(tx, state); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	//TODO on the response the idea was to redirect to a correct side, that has the response
	if authInfo.PollingCode {
		return &model.Response{
			Status:   fiber.StatusOK,
			Response: "Success! Please go back to the application that started this flow.", // TODO
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
	if stateInf.ResponseType == pkgModel.ResponseTypeTransferCode {
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
			Secure:   config.Get().Server.TLS.Enabled,
			HTTPOnly: true,
			SameSite: "Strict",
		}},
	}
}

func createSuperTokenEntry(tx *sqlx.Tx, authFlowInfo *authcodeinforepo.AuthFlowInfoOut, token *oauth2.Token, oidcSub string, networkData model.ClientMetaData) (*supertokenrepo.SuperTokenEntry, error) {
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
