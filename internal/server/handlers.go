package server

import "github.com/gofiber/fiber/v2"

func handleTest(ctx *fiber.Ctx) error {
	return ctx.SendString("This is a demo instance of the mytoken service. This service is currently under active development. For more information refer to: https://github.com/oidc-mytoken/server")
}
