package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/helmet/v2"

	loggerUtils "github.com/zachmann/mytoken/internal/utils/logger"
)

func addMiddlewares(s fiber.Router) {
	addRecoverMiddleware(s)
	addFaviconMiddleware(s)
	addLoggerMiddleware(s)
	addLimiterMiddleware(s)
	addHelmetMiddleware(s)
}

func addLoggerMiddleware(s fiber.Router) {
	s.Use(logger.New(logger.Config{
		Format:     "${time} ${ip} ${latency} - ${status} ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		Output:     loggerUtils.MustGetAccessLogger(),
	}))
}

func addLimiterMiddleware(s fiber.Router) {
	s.Use(limiter.New(limiter.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.IP() == "127.0.0.1"
		},
		Max:        100,
		Expiration: 5 * time.Minute,
	}))
}

func addCompressMiddleware(s fiber.Router) {
	s.Use(compress.New())
}

func addFaviconMiddleware(s fiber.Router) {
	s.Use(favicon.New(favicon.Config{
		File: "./web/static/favicon.ico",
	}))
}

func addRecoverMiddleware(s fiber.Router) {
	s.Use(recover.New())
}

func addHelmetMiddleware(s fiber.Router) {
	s.Use(helmet.New())
}
