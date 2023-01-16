package mytoken

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/config"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/polling"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken"
	"github.com/oidc-mytoken/server/internal/oidc/authcode"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleMytokenEndpoint handles requests on the mytoken endpoint
func HandleMytokenEndpoint(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	grantType, err := ctxutils.GetGrantType(ctx)
	if err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	rlog.WithField("grant_type", grantType.String()).Trace("Received mytoken request")
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
	res := model.Response{
		Status:   fiber.StatusBadRequest,
		Response: api.ErrorUnsupportedGrantType,
	}
	return res.Send(ctx)
}

func handleOIDCFlow(ctx *fiber.Ctx) error {
	req := response.NewOIDCFlowRequest()
	if err := json.Unmarshal(ctx.Body(), req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	_, ok := config.Get().ProviderByIssuer[req.Issuer]
	if !ok {
		return model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}.Send(ctx)
	}
	if len(req.Capabilities.Capabilities) == 0 {
		req.Capabilities.Capabilities = api.DefaultCapabilities
	}
	switch req.OIDCFlow {
	case model.OIDCFlowAuthorizationCode:
		authCodeReq := &response.AuthCodeFlowRequest{OIDCFlowRequest: *req}
		if err := json.Unmarshal(ctx.Body(), authCodeReq); err != nil {
			return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
		}
		return authcode.StartAuthCodeFlow(ctx, authCodeReq).Send(ctx)
	default:
		res := model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnsupportedOIDCFlow,
		}
		return res.Send(ctx)
	}
}
