package server

import (
	"fmt"
	"strings"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/mustache"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/endpoints"
	"github.com/oidc-mytoken/server/internal/endpoints/configuration"
	"github.com/oidc-mytoken/server/internal/endpoints/consent"
	"github.com/oidc-mytoken/server/internal/endpoints/redirect"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/server/routes"
	model2 "github.com/oidc-mytoken/server/pkg/model"
)

var server *fiber.App

var serverConfig = fiber.Config{
	ReadTimeout:    30 * time.Second,
	WriteTimeout:   90 * time.Second,
	IdleTimeout:    150 * time.Second,
	ReadBufferSize: 8192,
	// WriteBufferSize: 4096,
	ErrorHandler: handleError,
}

func initTemplateEngine() {
	engine := mustache.NewFileSystem(rice.MustFindBox("../../web").HTTPBox(), ".mustache")
	//TODO remove
	engine.Reload(true)
	serverConfig.Views = engine
}

// Init initializes the server
func Init() {
	initTemplateEngine()
	server = fiber.New(serverConfig)
	addMiddlewares(server)
	addRoutes(server)
	server.Use(func(ctx *fiber.Ctx) error {
		if len(ctx.Accepts(fiber.MIMETextHTML, fiber.MIMETextHTMLCharsetUTF8)) > 0 {
			return ctx.Render("sites/404", map[string]interface{}{
				"empty-navbar": true,
			}, "layouts/main")
		}
		return model.Response{
			Status: fiber.StatusNotFound,
			Response: model2.APIError{
				Error: "not_found",
			},
		}.Send(ctx)
	})
}

func addRoutes(s fiber.Router) {
	addWebRoutes(s)
	s.Get(routes.GetGeneralPaths().ConfigurationEndpoint, configuration.HandleConfiguration)
	s.Get("/.well-known/openid-configuration", func(ctx *fiber.Ctx) error {
		return ctx.Redirect(routes.GetGeneralPaths().ConfigurationEndpoint)
	})
	s.Get(routes.GetGeneralPaths().JWKSEndpoint, endpoints.HandleJWKS)
	s.Get(routes.GetGeneralPaths().OIDCRedirectEndpoint, redirect.HandleOIDCRedirect)
	s.Get("/c/:consent_code", consent.HandleConsent)
	s.Post("/c/:consent_code", consent.HandleConsentPost)
	addAPIRoutes(s)
}

func addWebRoutes(s fiber.Router) {
	s.Get("/", handleIndex)
	s.Get("/home", handleHome)
	s.Get("/native", handleNativeCallback)
}

func start(s *fiber.App) {
	if config.Get().Server.TLS.Enabled {
		if config.Get().Server.TLS.RedirectHTTP {
			httpServer := fiber.New(serverConfig)
			httpServer.All("*", func(ctx *fiber.Ctx) error {
				return ctx.Redirect(strings.Replace(ctx.Request().URI().String(), "http://", "https://", 1), fiber.StatusPermanentRedirect)
			})
			go httpServer.Listen(":80")
		}
		time.Sleep(time.Millisecond) // This is just for a more pretty output with the tls header printed after the http one
		log.Fatal(s.ListenTLS(":443", config.Get().Server.TLS.Cert, config.Get().Server.TLS.Key))
	} else {
		log.Fatal(s.Listen(fmt.Sprintf(":%d", config.Get().Server.Port)))
	}
}

// Start starts the server
func Start() {
	start(server)
}
