package server

import "github.com/gofiber/fiber/v2"

func handleTest(ctx *fiber.Ctx) error {
	return ctx.SendString("Hello")
}
