package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/utils/utils"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/endpoints"
	"github.com/oidc-mytoken/server/internal/endpoints/actions"
	"github.com/oidc-mytoken/server/internal/endpoints/configuration"
	"github.com/oidc-mytoken/server/internal/endpoints/consent"
	"github.com/oidc-mytoken/server/internal/endpoints/federation"
	"github.com/oidc-mytoken/server/internal/endpoints/notification/calendar"
	"github.com/oidc-mytoken/server/internal/endpoints/redirect"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/server/apipath"
	"github.com/oidc-mytoken/server/internal/server/paths"
	"github.com/oidc-mytoken/server/internal/server/spa"
	"github.com/oidc-mytoken/server/internal/server/ssh"
)

// essentialWebPaths are paths that must remain accessible even when web_interface is disabled.
// These include consent screens, native app callbacks, and legal pages.
var essentialWebPaths = []string{
	"/c/",
	"/native",
	"/privacy",
}

// IsEssentialWebPath checks if a path is an essential web path that must remain
// accessible even when the web interface is disabled.
func IsEssentialWebPath(path string) bool {
	for _, p := range essentialWebPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

var server *fiber.App

var serverConfig = fiber.Config{
	ReadTimeout:    30 * time.Second,
	WriteTimeout:   90 * time.Second,
	IdleTimeout:    150 * time.Second,
	ReadBufferSize: 8192,
	// WriteBufferSize: 4096,
	ErrorHandler: handleError,
	// ProxyHeader is set later from config
	Network: "tcp",
}

// Init initializes the server
func Init() {
	webEnabled := config.Get().Features.WebInterface.Enabled

	if !spa.Available {
		log.Fatal("SPA distribution not available. Please build the frontend first.")
	}

	// Set the SPA handler for consent page (needed even when web interface is disabled)
	consent.SPAHandler = spa.HandleSPAFallback()

	if webEnabled {
		log.Info("Web interface enabled")
	} else {
		log.Info("Web interface disabled (essential paths like consent and privacy remain accessible)")
	}

	serverConfig.ProxyHeader = config.Get().Server.ProxyHeader
	server = fiber.New(serverConfig)
	addMiddlewares(server)
	addRoutes(server)

	// Add SPA routes (needed for essential paths even when web interface is disabled)
	spa.AddRoutes(server)

	// Add fallback handler (404 for API, SPA fallback for web)
	server.Use(
		func(ctx *fiber.Ctx) error {
			path := ctx.Path()

			// For API routes, return JSON error
			if strings.HasPrefix(path, apipath.Prefix) {
				return jsonNotFound(ctx, path)
			}

			// For HTML requests
			if ctx.Accepts(fiber.MIMETextHTML, fiber.MIMETextHTMLCharsetUTF8) != "" {
				// Serve SPA for essential paths regardless of web interface setting
				if IsEssentialWebPath(path) {
					if handler := spa.HandleSPAFallback(); handler != nil {
						return handler(ctx)
					}
				}

				// If web interface is enabled, serve SPA for all HTML requests
				if webEnabled {
					if handler := spa.HandleSPAFallback(); handler != nil {
						return handler(ctx)
					}
				}
			}

			// Return JSON 404 for non-HTML requests or when web interface is disabled
			return jsonNotFound(ctx, path)
		},
	)
}

// jsonNotFound returns a JSON 404 error response
func jsonNotFound(ctx *fiber.Ctx, path string) error {
	return model.Response{
		Status: fiber.StatusNotFound,
		Response: api.Error{
			Error:            "not_found",
			ErrorDescription: path,
		},
	}.Send(ctx)
}

func addRoutes(s fiber.Router) {
	generalPaths := paths.GetGeneralPaths()
	s.Get(generalPaths.ConfigurationEndpoint, toFiberHandler(configuration.HandleConfiguration))
	s.Get(paths.WellknownOpenIDConfiguration, toFiberHandler(configuration.HandleConfiguration))
	if config.Get().Features.Federation.Enabled {
		s.Get(generalPaths.FederationEndpoint, federation.HandleEntityConfiguration)
	}
	s.Get(generalPaths.JWKSEndpoint, endpoints.HandleJWKS)
	s.Get(generalPaths.OIDCRedirectEndpoint, redirect.HandleOIDCRedirect)

	// Consent routes - these still need server-side handling even with SPA
	s.Get("/c/:consent_code", consent.HandleConsent)
	s.Post("/c/:consent_code", toFiberHandler(consent.HandleConsentPost))
	s.Post("/c", consent.HandleCreateConsent)

	// Native app callbacks - handled by SPA
	spaFallback := spa.HandleSPAFallback()
	s.Get("/native", spaFallback)
	s.Get("/native/abort", spaFallback)

	// Calendar ICS endpoint (always server-side)
	s.Get(utils.CombineURLPath(generalPaths.CalendarEndpoint, ":id"), calendar.HandleGetICS)

	s.Get(generalPaths.ActionsEndpoint, actions.HandleActions)
	addAPIRoutes(s)
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
