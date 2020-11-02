package super

import (
	"encoding/json"
	"fmt"

	"github.com/zachmann/mytoken/internal/config"

	"github.com/zachmann/mytoken/internal/endpoints/token/super/pkg"

	"github.com/zachmann/mytoken/internal/utils/ctxUtils"

	"github.com/gofiber/fiber/v2"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/oidc/authcode"
)

func HandleSuperTokenEndpoint(ctx *fiber.Ctx) error {
	grantType := ctxUtils.GetGrantType(ctx)
	switch grantType {
	case model.GrantTypeSuperToken:
		return fmt.Errorf("not yet implemented")
	case model.GrantTypeOIDCFlow:
		return handleOIDCFlow(ctx)
	case model.GrantTypeAccessToken:
		return fmt.Errorf("not yet implemented")
	case model.GrantTypePollingCode:
		return fmt.Errorf("not yet implemented")
	case model.GrantTypePrivateKeyJWT:
		return fmt.Errorf("not yet implemented")
	default:
		ctx.SendStatus(fiber.StatusBadRequest)
		return ctx.SendString("Bad grant_type")
	}
}

func handleOIDCFlow(ctx *fiber.Ctx) error {
	flow := ctxUtils.GetOIDCFlow(ctx)
	switch flow {
	case model.OIDCFlowAuthorizationCode:
		req := pkg.NewAuthCodeFlowRequest()
		if err := json.Unmarshal(ctx.Body(), &req); err != nil {
			return err
		}
		provider, ok := config.Get().ProviderByIssuer[req.Issuer]
		if !ok {
			ctx.SendStatus(fiber.StatusBadRequest)
			msg := fmt.Sprintf("Issuer '%s' not supported", req.Issuer)
			return ctx.SendString(msg)
		}
		ret, err := authcode.InitAuthCodeFlow(provider, req)
		if err != nil {
			return err
		}
		return ctx.JSON(ret)
	case model.OIDCFlowDevice:
		return fmt.Errorf("not yet implemented")
	default:
		ctx.SendStatus(fiber.StatusBadRequest)
		return ctx.SendString("Bad oidc_flow")
	}
}
