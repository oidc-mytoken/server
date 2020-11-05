package ctxUtils

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/zachmann/mytoken/internal/model"
)

type oidcFlowReqObj struct {
	OIDCFlow string `json:"oidc_flow"`
}

func GetOIDCFlowStr(ctx *fiber.Ctx) string {
	oidcFlow := ctx.Query("oidc_flow")
	if oidcFlow != "" {
		return oidcFlow
	}
	flow := oidcFlowReqObj{}
	err := json.Unmarshal(ctx.Body(), &flow)
	if err != nil {
		return ""
	}
	return flow.OIDCFlow
}

func GetOIDCFlow(ctx *fiber.Ctx) model.OIDCFlow {
	return model.NewOIDCFlow(GetOIDCFlowStr(ctx))
}
