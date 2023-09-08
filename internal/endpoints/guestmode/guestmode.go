package guestmode

import (
	"encoding/base64"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/utils/utils"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/server/paths"
	"github.com/oidc-mytoken/server/internal/server/routes"
	"github.com/oidc-mytoken/server/internal/utils/hashutils"
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

var stateDB map[string]string

func init() {
	stateDB = make(map[string]string)
}

func handleAuth(ctx *fiber.Ctx) error {
	state := ctx.Query("state")
	stateDB[state] = ctx.IP()
	return ctx.Redirect(routes.RedirectURI + "?state=" + state + "&code=" + state)
}

func handleToken(ctx *fiber.Ctx) error {
	var body struct {
		StateCode string `json:"code" xml:"code" form:"code"`
	}
	_ = ctx.BodyParser(&body)
	sub := "guest-"
	if body.StateCode != "" {
		sub += hashutils.SHA3_256Str([]byte(stateDB[body.StateCode]))
	} else {
		sub += hashutils.SHA3_256Str([]byte(utils.RandASCIIString(32)))
	}
	data, err := json.Marshal(map[string]string{"sub": sub})
	if err != nil {
		return err
	}
	idToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." + base64.RawURLEncoding.EncodeToString(data) + "."
	return ctx.JSON(
		map[string]any{
			"access_token":  utils.RandASCIIString(64),
			"refresh_token": utils.RandASCIIString(64),
			"id_token":      idToken,
			"expires_in":    600,
		},
	)
}
