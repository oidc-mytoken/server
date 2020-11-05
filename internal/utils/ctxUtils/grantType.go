package ctxUtils

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/zachmann/mytoken/internal/model"
)

type grantTypeReqObj struct {
	GrantType string `json:"grant_type"`
}

func GetGrantTypeStr(ctx *fiber.Ctx) string {
	grantType := ctx.Query("grant_type")
	if grantType != "" {
		return grantType
	}
	gt := grantTypeReqObj{}
	err := json.Unmarshal(ctx.Body(), &gt)
	if err != nil {
		return ""
	}
	return gt.GrantType
}

func GetGrantType(ctx *fiber.Ctx) model.GrantType {
	return model.NewGrantType(GetGrantTypeStr(ctx))
}
