package server

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/config"
	consent "github.com/oidc-mytoken/server/internal/endpoints/consent/pkg"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/templating"
)

func handleIndex(ctx *fiber.Ctx) error {
	binding := homeBindingData()
	binding[templating.MustacheKeyLoggedIn] = false
	binding[templating.MustacheKeyCookieLifetime] = cookies.CookieLifetime
	providers := []map[string]string{}
	for _, p := range config.Get().Providers {
		pp := make(map[string]string, 2)
		pp["issuer"] = p.Issuer
		pp["name"] = p.Name
		providers = append(providers, pp)
	}
	binding["providers"] = providers
	return ctx.Render("sites/home", binding, templating.LayoutMain)
}

func homeBindingData() map[string]interface{} {
	return map[string]interface{}{
		templating.MustacheKeyLoggedIn:             true,
		templating.MustacheKeyRestrictionsGUI:      true,
		templating.MustacheKeyHome:                 true,
		templating.MustacheKeyCapabilities:         consent.AllWebCapabilities(),
		templating.MustacheKeySubtokenCapabilities: consent.AllWebCapabilities(),
		templating.MustacheSubTokeninfo: map[string]interface{}{
			templating.MustacheKeyCollapse: templating.Collapsable{
				CollapseRestr: true,
			},
			templating.MustacheKeyPrefix:   "tokeninfo-",
			templating.MustacheKeyReadOnly: true,
		},
		templating.MustacheSubCreateMT: map[string]interface{}{
			templating.MustacheKeyPrefix: "createMT-",
		},
	}
}

func handleHome(ctx *fiber.Ctx) error {
	return ctx.Render("sites/home", homeBindingData(), templating.LayoutMain)
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
				templating.MustacheKeyRestrictionsGUI:      true,
				templating.MustacheKeyRestrictions:         consent.WebRestrictions{},
				templating.MustacheKeyCapabilities:         consent.AllWebCapabilities(),
				templating.MustacheKeySubtokenCapabilities: consent.AllWebCapabilities(),
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
		templating.MustacheKeyGrants:               grants,
		templating.MustacheKeyLoggedIn:             true,
		templating.MustacheKeySettings:             true,
		templating.MustacheKeySettingsSSH:          true,
		templating.MustacheKeyRestrictionsGUI:      true,
		templating.MustacheKeyRestrictions:         consent.WebRestrictions{},
		templating.MustacheKeyCapabilities:         consent.AllWebCapabilities(),
		templating.MustacheKeySubtokenCapabilities: consent.AllWebCapabilities(),
	}
	return ctx.Render("sites/settings", binding, templating.LayoutMain)
}

func handleNativeCallback(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		templating.MustacheKeyEmptyNavbar: true,
		templating.MustacheKeyApplication: ctx.Query("application"),
	}
	return ctx.Render("sites/native", binding, templating.LayoutMain)
}

func handleNativeConsentAbortCallback(ctx *fiber.Ctx) error {
	binding := map[string]interface{}{
		templating.MustacheKeyEmptyNavbar: true,
		templating.MustacheKeyApplication: ctx.Query("application"),
	}
	return ctx.Render("sites/native.abort", binding, templating.LayoutMain)
}

func handlePrivacy(ctx *fiber.Ctx) error {
	so := config.Get().ServiceOperator
	binding := map[string]interface{}{
		templating.MustacheKeyEmptyNavbar:    true,
		templating.MustacheKeyName:           so.Name,
		templating.MustacheKeyHomepage:       so.Homepage,
		templating.MustacheKeyContact:        so.Contact,
		templating.MustacheKeyPrivacyContact: so.Privacy,
	}
	return ctx.Render("sites/privacy", binding, templating.LayoutMain)
}
