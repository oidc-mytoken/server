package guestmode

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/utils/utils"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/server/paths"
	"github.com/oidc-mytoken/server/internal/server/routes"
)

func Init(s fiber.Router) {
	if !config.Get().Features.GuestMode.Enabled {
		return
	}
	baseURL := paths.GetCurrentAPIPaths().GuestModeOP
	conf = map[string]any{
		"token_endpoint":         utils.CombineURLPath(config.Get().IssuerURL, baseURL, "token"),
		"authorization_endpoint": utils.CombineURLPath(config.Get().IssuerURL, baseURL, "auth"),
	}
	router := s.Group(baseURL)
	router.Get(paths.WellknownOpenIDConfiguration, handleConfig)
	router.Get("auth", handleAuth)
	router.Post("token", handleToken)
}

var conf map[string]any

func handleConfig(ctx *fiber.Ctx) error {
	return ctx.JSON(conf)
}

func handleAuth(ctx *fiber.Ctx) error {
	state := ctx.Query("state")
	return ctx.Redirect(routes.RedirectURI + "?state=" + state + "&code=code")
}

func handleToken(ctx *fiber.Ctx) error {
	return ctx.JSON(
		map[string]any{
			"access_token":  utils.RandASCIIString(64),
			"refresh_token": utils.RandASCIIString(64),
			"id_token": `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJndWVzdCJ9.
OI5skE5VAlQjI4rqAFUjqwGyEnmmQNXBTOvO7pukZoo`,
			"expires_in": 600,
		},
	)
}
