package ctxUtils

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/server/shared/mytoken/universalmytoken"
)

// GetMytokenStr checks a fiber.Ctx for a mytoken and returns the token string as passed to the request
func GetMytokenStr(ctx *fiber.Ctx) string {
	req := api.CreateTransferCodeRequest{}
	if err := json.Unmarshal(ctx.Body(), &req); err == nil {
		if req.Mytoken != "" {
			return req.Mytoken
		}
	}
	if tok := GetAuthHeaderToken(ctx); tok != "" {
		return tok
	}
	if tok := ctx.Cookies("mytoken"); tok != "" {
		return tok
	}
	return ""
}

// GetMytoken checks a fiber.Ctx for a mytoken and returns a token object
func GetMytoken(ctx *fiber.Ctx) *universalmytoken.UniversalMytoken {
	tok := GetMytokenStr(ctx)
	if tok == "" {
		return nil
	}
	t, err := universalmytoken.Parse(tok)
	if err != nil {
		return nil
	}
	return &t
}
