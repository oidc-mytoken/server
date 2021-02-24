package ctxUtils

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/endpoints/token/super/pkg"
	"github.com/oidc-mytoken/server/shared/supertoken/token"
)

// GetSuperTokenStr checks a fiber.Ctx for a super token and returns the token string as passed to the request
func GetSuperTokenStr(ctx *fiber.Ctx) string {
	req := pkg.CreateTransferCodeRequest{}
	if err := json.Unmarshal(ctx.Body(), &req); err == nil {
		if req.SuperToken != "" {
			return req.SuperToken
		}
	}
	if tok := GetAuthHeaderToken(ctx); tok != "" {
		return tok
	}
	if tok := ctx.Cookies("mytoken-supertoken"); tok != "" {
		return tok
	}
	return ""
}

// GetSuperToken checks a fiber.Ctx for a super token and returns a token object
func GetSuperToken(ctx *fiber.Ctx) *token.Token {
	tok := GetSuperTokenStr(ctx)
	if tok == "" {
		return nil
	}
	t, err := token.GetLongSuperToken(tok)
	if err != nil {
		return nil
	}
	return &t
}
