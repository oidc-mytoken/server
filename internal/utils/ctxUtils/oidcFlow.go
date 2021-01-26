package ctxUtils

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/pkg/model"
)

type oidcFlowReqObj struct {
	OIDCFlow string `json:"oidc_flow"`
}

// GetOIDCFlowStr returns the oidc flow string for a fiber.Ctx by checking the query as well as the request body (json)
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

// GetOIDCFlow returns the model.OIDCFlow for a fiber.Ctx by checking the query as well as the request body (json)
func GetOIDCFlow(ctx *fiber.Ctx) model.OIDCFlow {
	return model.NewOIDCFlow(GetOIDCFlowStr(ctx))
}
