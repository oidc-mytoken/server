package server

import (
	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/model"
)

type handler func(ctx *fiber.Ctx) *model.Response

func toFiberHandler(h handler) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return h(ctx).Send(ctx)
	}
}
