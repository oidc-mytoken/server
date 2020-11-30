package ctxUtils

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/zachmann/mytoken/internal/model"
)

type grantTypeReqObj struct {
	GrantType string `json:"grant_type"`
}

func GetGrantTypeStr(ctx *fiber.Ctx) (string, error) {
	grantType := ctx.Query("grant_type")
	if grantType != "" {
		return grantType, nil
	}
	gt := grantTypeReqObj{}
	err := json.Unmarshal(ctx.Body(), &gt)
	if err != nil {
		return "", err
	}
	return gt.GrantType, nil
}

func GetGrantType(ctx *fiber.Ctx) (model.GrantType, error) {
	gt, err := GetGrantTypeStr(ctx)
	return model.NewGrantType(gt), err
}
