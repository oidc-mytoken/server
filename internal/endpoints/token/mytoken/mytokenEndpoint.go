package mytoken

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/polling"
	serverModel "github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/authcode"
	"github.com/oidc-mytoken/server/internal/utils/ctxUtils"
	"github.com/oidc-mytoken/server/pkg/api/v0"
	"github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/mytoken"
)

// HandleMytokenEndpoint handles requests on the mytoken endpoint
func HandleMytokenEndpoint(ctx *fiber.Ctx) error {
	grantType, err := ctxUtils.GetGrantType(ctx)
	if err != nil {
		return serverModel.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	log.WithField("grant_type", grantType).Trace("Received mytoken request")
	switch grantType {
	case model.GrantTypeMytoken:
		return mytoken.HandleMytokenFromMytoken(ctx).Send(ctx)
	case model.GrantTypeOIDCFlow:
		return handleOIDCFlow(ctx)
	case model.GrantTypePollingCode:
		if config.Get().Features.Polling.Enabled {
			return polling.HandlePollingCode(ctx)
		}
	case model.GrantTypeAccessToken:
		if config.Get().Features.AccessTokenGrant.Enabled {
			return serverModel.ResponseNYI.Send(ctx)
		}
	case model.GrantTypePrivateKeyJWT:
		if config.Get().Features.SignedJWTGrant.Enabled {
			return serverModel.ResponseNYI.Send(ctx)
		}
	case model.GrantTypeTransferCode:
		if config.Get().Features.TransferCodes.Enabled {
			return mytoken.HandleMytokenFromTransferCode(ctx).Send(ctx)
		}
	}
	res := serverModel.Response{
		Status:   fiber.StatusBadRequest,
		Response: api.APIErrorUnsupportedGrantType,
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
			Response: api.APIErrorUnknownIssuer,
		}.Send(ctx)
	}
	switch req.OIDCFlow {
	case model.OIDCFlowAuthorizationCode:
		return authcode.StartAuthCodeFlow(ctx, *req).Send(ctx)
	// case model.OIDCFlowDevice:
	// 	return serverModel.ResponseNYI.Send(ctx)
	default:
		res := serverModel.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.APIErrorUnsupportedOIDCFlow,
		}
		return res.Send(ctx)
	}
}
