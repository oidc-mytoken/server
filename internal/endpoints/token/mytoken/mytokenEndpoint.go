package mytoken

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/config"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/polling"
	serverModel "github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/authcode"
	"github.com/oidc-mytoken/server/internal/utils/ctxUtils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/mytoken"
)

var defaultCapabilities = api.Capabilities{
	api.CapabilityAT,
	api.CapabilityTokeninfo,
}

// HandleMytokenEndpoint handles requests on the mytoken endpoint
func HandleMytokenEndpoint(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	grantType, err := ctxUtils.GetGrantType(ctx)
	if err != nil {
		return serverModel.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	rlog.WithField("grant_type", grantType).Trace("Received mytoken request")
	switch grantType {
	case model.GrantTypeMytoken:
		return mytoken.HandleMytokenFromMytoken(ctx).Send(ctx)
	case model.GrantTypeOIDCFlow:
		return handleOIDCFlow(ctx)
	case model.GrantTypePollingCode:
		if config.Get().Features.Polling.Enabled {
			return polling.HandlePollingCode(ctx)
		}
	case model.GrantTypeTransferCode:
		if config.Get().Features.TransferCodes.Enabled {
			return mytoken.HandleMytokenFromTransferCode(ctx).Send(ctx)
		}
	}
	res := serverModel.Response{
		Status:   fiber.StatusBadRequest,
		Response: api.ErrorUnsupportedGrantType,
	}
	return res.Send(ctx)
}

func handleOIDCFlow(ctx *fiber.Ctx) error {
	req := response.NewOIDCFlowRequest()
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return serverModel.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	_, ok := config.Get().ProviderByIssuer[req.Issuer]
	if !ok {
		return serverModel.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}.Send(ctx)
	}
	if len(req.Capabilities) == 0 {
		req.Capabilities = defaultCapabilities
	}
	switch req.OIDCFlow {
	case model.OIDCFlowAuthorizationCode:
		return authcode.StartAuthCodeFlow(ctx, req).Send(ctx)
	// case model.OIDCFlowDevice:
	// 	return serverModel.ResponseNYI.Send(ctx)
	default:
		res := serverModel.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnsupportedOIDCFlow,
		}
		return res.Send(ctx)
	}
}
