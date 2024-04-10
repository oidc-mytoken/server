package notifier

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/notifier/pkg"
)

var server *fiber.App

var serverConfig = fiber.Config{
	ReadTimeout:    30 * time.Second,
	WriteTimeout:   90 * time.Second,
	IdleTimeout:    150 * time.Second,
	ReadBufferSize: 32768,
}

func startServer() {
	server = fiber.New(serverConfig)
	server.Use(recover.New())
	server.Use(helmet.New())
	server.Use(requestid.New())

	server.Post(
		ServerPaths.Email, func(ctx *fiber.Ctx) error {
			var req pkg.EmailNotificationRequest
			if err := ctx.BodyParser(&req); err != nil {
				return err
			}
			if err := HandleEmailRequest(req); err != nil {
				return err
			}
			return ctx.Status(fiber.StatusNoContent).Send(nil)
		},
	)
	log.WithError(server.Listen(":40111")).Fatal()
}
