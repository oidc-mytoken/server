package ctxUtils

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/shared/model"
)

type oidcFlowReqObj struct {
	OIDCFlow string `json:"oidc_flow"`
}

// GetOIDCFlowStr returns the oidc flow string for a fiber.Ctx by checking the query as well as the request body (json)
func GetOIDCFlowStr(ctx *fiber.Ctx) (string, error) {
	oidcFlow := ctx.Query("oidc_flow")
	if oidcFlow != "" {
		return oidcFlow, nil
	}
	flow := oidcFlowReqObj{}
	err := json.Unmarshal(ctx.Body(), &flow)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return flow.OIDCFlow, nil
}

// GetOIDCFlow returns the model.OIDCFlow for a fiber.Ctx by checking the query as well as the request body (json)
func GetOIDCFlow(ctx *fiber.Ctx) (model.OIDCFlow, error) {
	f, err := GetOIDCFlowStr(ctx)
	return model.NewOIDCFlow(f), err
}
