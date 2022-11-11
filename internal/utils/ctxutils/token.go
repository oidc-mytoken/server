package ctxutils

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/shared/mytoken/universalmytoken"
)

// GetMytokenStr checks a fiber.Ctx for a mytoken and returns the token string as passed to the request
func GetMytokenStr(ctx *fiber.Ctx) string {
	req := struct {
		Mytoken string `json:"mytoken"`
	}{}
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
func GetMytoken(ctx *fiber.Ctx) (*universalmytoken.UniversalMytoken, bool) {
	rlog := logger.GetRequestLogger(ctx)
	tok := GetMytokenStr(ctx)
	if tok == "" {
		return nil, false
	}
	t, err := universalmytoken.Parse(rlog, tok)
	if err != nil {
		return nil, true
	}
	return &t, true
}
