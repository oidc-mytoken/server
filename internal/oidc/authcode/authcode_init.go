package authcode

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/unixtime"
	"github.com/oidc-mytoken/utils/utils"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/oidc/pkce"
	provider2 "github.com/oidc-mytoken/server/internal/oidc/provider"
	"github.com/oidc-mytoken/server/internal/server/httpstatus"
	"github.com/oidc-mytoken/server/internal/server/routes"
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
