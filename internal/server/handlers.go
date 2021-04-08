package server

import (
	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/config"
)

func handleIndex(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		loggedIn: false,
	}
	providers := []map[string]string{}
	for _, p := range config.Get().Providers {
		pp := make(map[string]string, 2)
		pp["issuer"] = p.Issuer
		pp["name"] = p.Name
		providers = append(providers, pp)
	}
	binding["providers"] = providers
	return ctx.Render("sites/index", binding, layoutMain)
}

func handleHome(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		loggedIn: true,
	}
	return ctx.Render("sites/home", binding, layoutMain)
}

func handleNativeCallback(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		emptyNavbar: true,
	}
	return ctx.Render("sites/native", binding, layoutMain)
}

func handleNativeConsentAbortCallback(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		emptyNavbar: true,
	}
	return ctx.Render("sites/native.abort", binding, layoutMain)
}

func handlePrivacy(ctx *fiber.Ctx) error {
	so := config.Get().ServiceOperator
	binding := map[string]interface{}{
		emptyNavbar:       true,
		"name":            so.Name,
		"homepage":        so.Homepage,
		"contact":         so.Contact,
		"privacy-contact": so.Privacy,
	}
	return ctx.Render("sites/privacy", binding, layoutMain)
}
