package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/helmet/v2"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/server/apiPath"
	"github.com/oidc-mytoken/server/internal/server/routes"
	"github.com/oidc-mytoken/server/internal/utils/fileio"
	loggerUtils "github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/shared/utils"
)

//go:embed web/static
var _staticFS embed.FS
var staticFS fs.FS

//go:embed web/static/img/favicon.ico
var _faviconFS embed.FS
var faviconFS fs.FS

func init() {
	var err error
	staticFS, err = fs.Sub(_staticFS, "web/static")
	if err != nil {
		log.WithError(err).Fatal()
	}
	faviconFS, err = fs.Sub(_faviconFS, "web/static/img")
	if err != nil {
		log.WithError(err).Fatal()
	}
}

func addMiddlewares(s fiber.Router) {
	addRecoverMiddleware(s)
	addRequestIDMiddleware(s)
	addLoggerMiddleware(s)
	addLimiterMiddleware(s)
	addCorsMiddleware(s)
	addFaviconMiddleware(s)
	addHelmetMiddleware(s)
	addStaticFiles(s)
	addCompressMiddleware(s)
}

func addLoggerMiddleware(s fiber.Router) {
	s.Use(
		logger.New(
			logger.Config{
				Format:     "${time} ${ip} ${ua} ${latency} - ${status} ${method} ${path} ${locals:requestid}\n",
				TimeFormat: "2006-01-02 15:04:05",
				Output:     loggerUtils.MustGetAccessLogger(),
			},
		),
	)
}

func addLimiterMiddleware(s fiber.Router) {
	limiterConf := config.Get().Server.Limiter
	if !limiterConf.Enabled {
		return
	}
	s.Use(
		limiter.New(
			limiter.Config{
				Next: func(c *fiber.Ctx) bool {
					return utils.IPIsIn(c.IP(), limiterConf.AlwaysAllow)
				},
				Max:        limiterConf.Max,
				Expiration: time.Duration(limiterConf.Window) * time.Second,
			},
		),
	)
}

func addCompressMiddleware(s fiber.Router) {
	s.Use(compress.New())
}

func addStaticFiles(s fiber.Router) {
	s.Use(
		"/static", filesystem.New(
			filesystem.Config{
				Root: fileio.NewLocalAndOtherSearcherFilesystem(
					fileio.JoinIfFirstNotEmpty(
						config.Get().Features.WebInterface.OverwriteDir, "static",
					), http.FS(staticFS),
				),
				MaxAge: 3600,
			},
		),
	)
}

func addFaviconMiddleware(s fiber.Router) {
	s.Use(
		favicon.New(
			favicon.Config{
				File: "favicon.ico",
				FileSystem: fileio.NewLocalAndOtherSearcherFilesystem(
					fileio.JoinIfFirstNotEmpty(
						config.Get().Features.WebInterface.OverwriteDir, "static/img",
					), http.FS(faviconFS),
				),
			},
		),
	)
}

func addRecoverMiddleware(s fiber.Router) {
	s.Use(recover.New())
}

func addHelmetMiddleware(s fiber.Router) {
	s.Use(helmet.New())
}

func addRequestIDMiddleware(s fiber.Router) {
	s.Use(requestid.New())
}

func addCorsMiddleware(s fiber.Router) {
	allowedPaths := []string{
		routes.WellknownMytokenConfiguration,
		routes.WellknownOpenIDConfiguration,
	}
	allowedPrefixes := []string{
		apiPath.Prefix,
		"/static",
	}
	s.Use(
		cors.New(
			cors.Config{
				Next: func(c *fiber.Ctx) bool {
					p := c.Path()
					for _, pre := range allowedPrefixes {
						if strings.HasPrefix(p, pre) {
							return false
						}
					}
					for _, pre := range allowedPaths {
						if p == pre {
							return false
						}
					}
					return true
				},
			},
		),
	)
}
