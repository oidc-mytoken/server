package ctxUtils

import (
	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/model"
)

// ClientMetaData returns the model.ClientMetaData for a given fiber.Ctx
func ClientMetaData(ctx *fiber.Ctx) *model.ClientMetaData {
	return &model.ClientMetaData{
		IP:        ctx.IP(),
		UserAgent: string(ctx.Request().Header.UserAgent()),
	}
}
