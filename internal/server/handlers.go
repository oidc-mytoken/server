package server

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"

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
	type bindData struct {
		DisplayName string
		Name        string
		Description string
		Link        string
		EmbedBody   string
		partialName string
		bindingData map[string]interface{}
	}
	grants := []*bindData{
		{
			DisplayName: "SSH",
			Name:        "ssh",
			Description: "The SSH grant type allows you to link an ssh key and use ssh authentication for various actions.",
			Link:        "/settings/grants/ssh",
			partialName: "sites/settings-ssh",
			bindingData: map[string]interface{}{
				"restr-gui":             true,
				"restrictions":          consent.WebRestrictions{},
				"capabilities":          consent.AllWebCapabilities(),
				"subtoken-capabilities": consent.AllWebCapabilities(),
			},
		},
	}
	for _, g := range grants {
		embed := strings.Builder{}
		if err := errors.WithStack(serverConfig.Views.Render(&embed, g.partialName, g.bindingData)); err != nil {
			return err
		}
		g.EmbedBody = embed.String()
	}
	binding := map[string]interface{}{
		"grants":                grants,
		loggedIn:                true,
		"settings":              true,
		"settings-ssh":          true,
		"restr-gui":             true,
		"restrictions":          consent.WebRestrictions{},
		"capabilities":          consent.AllWebCapabilities(),
		"subtoken-capabilities": consent.AllWebCapabilities(),
	}
	return ctx.Render("sites/settings", binding, layoutMain)
}

func handleNativeCallback(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		emptyNavbar:   true,
		"application": ctx.Query("application"),
	}
	return ctx.Render("sites/native", binding, layoutMain)
}

func handleNativeConsentAbortCallback(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		emptyNavbar:   true,
		"application": ctx.Query("application"),
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
