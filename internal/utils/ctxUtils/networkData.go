package ctxUtils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zachmann/mytoken/internal/model"
)

func NetworkData(ctx *fiber.Ctx) *model.NetworkData {
	return &model.NetworkData{
		IP:        ctx.IP(),
		UserAgent: string(ctx.Request().Header.UserAgent()),
	}
}
