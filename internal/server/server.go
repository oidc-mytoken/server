package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/mustache"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/endpoints"
	"github.com/oidc-mytoken/server/internal/endpoints/configuration"
	"github.com/oidc-mytoken/server/internal/endpoints/consent"
	"github.com/oidc-mytoken/server/internal/endpoints/redirect"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/server/apiPath"
	"github.com/oidc-mytoken/server/internal/server/routes"
	"github.com/oidc-mytoken/server/internal/server/ssh"
)

var server *fiber.App

var serverConfig = fiber.Config{
	ReadTimeout:    30 * time.Second,
	WriteTimeout:   90 * time.Second,
	IdleTimeout:    150 * time.Second,
	ReadBufferSize: 8192,
	// WriteBufferSize: 4096,
	ErrorHandler: handleError,
	// ProxyHeader is set later from config
}

//go:embed web/sites web/layouts
var _webFiles embed.FS
var webFiles fs.FS

//go:embed web/partials
var _partials embed.FS
var partials fs.FS

func init() {
	var err error
	webFiles, err = fs.Sub(_webFiles, "web")
	if err != nil {
		log.WithError(err).Fatal()
	}
	partials, err = fs.Sub(_partials, "web/partials")
	if err != nil {
		log.WithError(err).Fatal()
	}
}

func initTemplateEngine() {
	engine := mustache.NewFileSystemPartials(http.FS(webFiles), ".mustache", http.FS(partials))
	serverConfig.Views = engine
}

// Init initializes the server
func Init() {
	initTemplateEngine()
	serverConfig.ProxyHeader = config.Get().Server.ProxyHeader
	server = fiber.New(serverConfig)
	addMiddlewares(server)
	addRoutes(server)
	server.Use(
		func(ctx *fiber.Ctx) error {
			path := ctx.Path()
			if !strings.HasPrefix(path, apiPath.Prefix) && ctx.Accepts(
				fiber.MIMETextHTML, fiber.MIMETextHTMLCharsetUTF8,
			) != "" {
				ctx.Status(fiber.StatusNotFound)
				return ctx.Render(
					"sites/404", map[string]interface{}{
						"empty-navbar": true,
					}, "layouts/main",
				)
			}
			return model.Response{
				Status: fiber.StatusNotFound,
				Response: api.Error{
					Error:            "not_found",
					ErrorDescription: path,
				},
			}.Send(ctx)
		},
	)
}

func addRoutes(s fiber.Router) {
	addWebRoutes(s)
	s.Get(routes.GetGeneralPaths().ConfigurationEndpoint, configuration.HandleConfiguration)
	s.Get("/.well-known/openid-configuration", configuration.HandleConfiguration)
	s.Get(routes.GetGeneralPaths().JWKSEndpoint, endpoints.HandleJWKS)
	s.Get(routes.GetGeneralPaths().OIDCRedirectEndpoint, redirect.HandleOIDCRedirect)
	s.Get("/c/:consent_code", consent.HandleConsent)
	s.Post("/c/:consent_code", consent.HandleConsentPost)
	s.Get("/native", handleNativeCallback)
	s.Get("/native/abort", handleNativeConsentAbortCallback)
	s.Get(routes.GetGeneralPaths().Privacy, handlePrivacy)
	s.Get("/settings", handleSettings)
	s.Get("/settings/grants/ssh", handleSSH)
	addAPIRoutes(s)
}

func addWebRoutes(s fiber.Router) {
	s.Get("/", handleIndex)
	s.Get("/home", handleHome)
}

func start(s *fiber.App) {
	if config.Get().Features.SSH.Enabled {
		go ssh.Serve()
	}
	if !config.Get().Server.TLS.Enabled {
		log.WithField("port", config.Get().Server.Port).Info("TLS is disabled starting http server")
		log.WithError(s.Listen(fmt.Sprintf(":%d", config.Get().Server.Port))).Fatal()
	}
	// TLS enabled
	if config.Get().Server.TLS.RedirectHTTP {
		httpServer := fiber.New(serverConfig)
		httpServer.All(
			"*", func(ctx *fiber.Ctx) error {
				//goland:noinspection HttpUrlsUsage
				return ctx.Redirect(
					strings.Replace(ctx.Request().URI().String(), "http://", "https://", 1),
					fiber.StatusPermanentRedirect,
				)
			},
		)
		log.Info("TLS and http redirect enabled, starting redirect server on port 80")
		go func() {
			log.WithError(httpServer.Listen(":80")).Fatal()
		}()
	}
	time.Sleep(time.Millisecond) // This is just for a more pretty output with the tls header printed after the http one
	log.Info("TLS enabled, starting https server on port 443")
	log.WithError(s.ListenTLS(":443", config.Get().Server.TLS.Cert, config.Get().Server.TLS.Key)).Fatal()
}

// Start starts the server
func Start() {
	start(server)
}
