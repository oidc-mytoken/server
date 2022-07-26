package server

import (
	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/config"
	consent "github.com/oidc-mytoken/server/internal/endpoints/consent/pkg"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
)

func handleIndex(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		loggedIn:          false,
		"cookie-lifetime": cookies.CookieLifetime,
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
		loggedIn:                true,
		restrictionsGUI:         true,
		"home":                  true,
		"capabilities":          consent.AllWebCapabilities(),
		"subtoken-capabilities": consent.AllWebCapabilities(),
	}
	return ctx.Render("sites/home", binding, layoutMain)
}

func handleSettings(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		loggedIn:   true,
		"settings": true,
		"grants": []struct {
			DisplayName string
			Name        string
			Description string
			Link        string
		}{
			{
				DisplayName: "SSH",
				Name:        "ssh",
				Description: "The SSH grant type allows you to link an ssh key and use ssh authentication for various actions.",
				Link:        "/settings/grants/ssh",
			},
		},
	}
	return ctx.Render("sites/settings", binding, layoutMain)
}
func handleSSH(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		"settings-ssh":          true,
		loggedIn:                true,
		"restr-gui":             true,
		"restrictions":          consent.WebRestrictions{},
		"capabilities":          consent.AllWebCapabilities(),
		"subtoken-capabilities": consent.AllWebCapabilities(),
	}
	return ctx.Render("sites/ssh", binding, layoutMain)
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
