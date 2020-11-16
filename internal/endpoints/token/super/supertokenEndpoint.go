package super

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zachmann/mytoken/internal/endpoints/token/super/polling"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/oidc/authcode"
	"github.com/zachmann/mytoken/internal/utils/ctxUtils"
)

func HandleSuperTokenEndpoint(ctx *fiber.Ctx) error {
	grantType := ctxUtils.GetGrantType(ctx)
	switch grantType {
	case model.GrantTypeSuperToken:
		return model.ResponseNYI.Send(ctx)
	case model.GrantTypeOIDCFlow:
		return handleOIDCFlow(ctx)
	case model.GrantTypePollingCode:
		return polling.HandlePollingCode(ctx)
	case model.GrantTypeAccessToken:
		return model.ResponseNYI.Send(ctx)
	case model.GrantTypePrivateKeyJWT:
		return model.ResponseNYI.Send(ctx)
	default:
		res := model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.APIErrorUnknownGrantType,
		}
		return res.Send(ctx)
	}
}

func handleOIDCFlow(ctx *fiber.Ctx) error {
	flow := ctxUtils.GetOIDCFlow(ctx)
	switch flow {
	case model.OIDCFlowAuthorizationCode:
		res := authcode.InitAuthCodeFlow(ctx.Body())
		return res.Send(ctx)
	case model.OIDCFlowDevice:
		return model.ResponseNYI.Send(ctx)
	default:
		res := model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.APIErrorUnknownOIDCFlow,
		}
		return res.Send(ctx)
	}
}
