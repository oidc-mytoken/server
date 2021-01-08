package super

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/server/config"
	"github.com/zachmann/mytoken/internal/server/endpoints/token/super/polling"
	serverModel "github.com/zachmann/mytoken/internal/server/model"
	"github.com/zachmann/mytoken/internal/server/oidc/authcode"
	"github.com/zachmann/mytoken/internal/server/supertoken"
	"github.com/zachmann/mytoken/internal/server/utils/ctxUtils"
	"github.com/zachmann/mytoken/pkg/model"
)

// HandleSuperTokenEndpoint handles requests on the super token endpoint
func HandleSuperTokenEndpoint(ctx *fiber.Ctx) error {
	grantType, err := ctxUtils.GetGrantType(ctx)
	if err != nil {
		return serverModel.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	log.WithField("grant_type", grantType).Trace("Received super token request")
	switch grantType {
	case model.GrantTypeSuperToken:
		return supertoken.HandleSuperTokenFromSuperToken(ctx).Send(ctx)
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
			return supertoken.HandleSuperTokenFromTransferCode(ctx).Send(ctx)
		}
	}
	res := serverModel.Response{
		Status:   fiber.StatusBadRequest,
		Response: model.APIErrorUnsupportedGrantType,
	}
	return res.Send(ctx)
}

func handleOIDCFlow(ctx *fiber.Ctx) error {
	flow := ctxUtils.GetOIDCFlow(ctx)
	switch flow {
	case model.OIDCFlowAuthorizationCode:
		return authcode.StartAuthCodeFlow(ctx.Body()).Send(ctx)
	case model.OIDCFlowDevice:
		return serverModel.ResponseNYI.Send(ctx)
	default:
		res := serverModel.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.APIErrorUnsupportedOIDCFlow,
		}
		return res.Send(ctx)
	}
}
