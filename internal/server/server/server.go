package server

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/server/config"
	"github.com/zachmann/mytoken/internal/server/endpoints"
	"github.com/zachmann/mytoken/internal/server/endpoints/configuration"
	"github.com/zachmann/mytoken/internal/server/endpoints/redirect"
	"github.com/zachmann/mytoken/internal/server/server/routes"
)

var server *fiber.App

// Init initializes the server
func Init() {
	server = fiber.New(fiber.Config{
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   90 * time.Second,
		IdleTimeout:    150 * time.Second,
		ReadBufferSize: 8192,
		// WriteBufferSize: 4096,
	})
	addMiddlewares(server)
	addRoutes(server)
}

func addRoutes(s fiber.Router) {
	s.Get("/", handleTest)
	s.Get(routes.GetGeneralPaths().ConfigurationEndpoint, configuration.HandleConfiguration)
	s.Get("/.well-known/openid-configuration", func(ctx *fiber.Ctx) error {
		return ctx.Redirect(routes.GetGeneralPaths().ConfigurationEndpoint)
	})
	s.Get(routes.GetGeneralPaths().JWKSEndpoint, endpoints.HandleJWKS)
	s.Get(routes.GetGeneralPaths().OIDCRedirectEndpoint, redirect.HandleOIDCRedirect)
	addAPIRoutes(s)
}

func start(s *fiber.App) {
	log.Fatal(s.Listen(fmt.Sprintf(":%d", config.Get().Server.Port)))
}

// Start starts the server
func Start() {
	start(server)
}
