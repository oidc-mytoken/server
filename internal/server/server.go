package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/endpoints"
	"github.com/zachmann/mytoken/internal/endpoints/configuration"
	"github.com/zachmann/mytoken/internal/endpoints/nativeredirect"
	"github.com/zachmann/mytoken/internal/endpoints/redirect"
	"github.com/zachmann/mytoken/internal/server/routes"
	"github.com/zachmann/mytoken/shared/utils"
)

var server *fiber.App

var serverConfig = fiber.Config{
	ReadTimeout:    30 * time.Second,
	WriteTimeout:   90 * time.Second,
	IdleTimeout:    150 * time.Second,
	ReadBufferSize: 8192,
	// WriteBufferSize: 4096,
}

// Init initializes the server
func Init() {
	server = fiber.New(serverConfig)
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
	s.Get(utils.CombineURLPath(routes.GetGeneralPaths().NativeRedirectEndpoint, ":poll"), nativeredirect.HandleNativeRedirect)
	addAPIRoutes(s)
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
