package server

import (
	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/config"
)

func handleIndex(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		"logged-in": false,
	}
	providers := []map[string]string{}
	for _, p := range config.Get().Providers {
		pp := make(map[string]string, 2)
		pp["issuer"] = p.Issuer
		pp["name"] = p.Name
		providers = append(providers, pp)
	}
	binding["providers"] = providers
	return ctx.Render("sites/index", binding, "layouts/main")
}

func handleHome(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		"logged-in": true,
	}
	return ctx.Render("sites/home", binding, "layouts/main")
}

func handleNativeCallback(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		"empty-navbar": true,
	}
	return ctx.Render("sites/native", binding, "layouts/main")
}

func handleNativeConsentAbortCallback(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		"empty-navbar": true,
	}
	return ctx.Render("sites/native.abort", binding, "layouts/main")
}
