package ctxutils

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

func RenderErrorPage(ctx *fiber.Ctx, status int, errorMsg string, optionalErrorHeading ...string) error {
	errorHeading := http.StatusText(status)
	if len(optionalErrorHeading) > 0 {
		errorHeading = optionalErrorHeading[0]
	}
	return ctx.Status(status).Render(
		"sites/error", map[string]interface{}{
			"empty-navbar":  true,
			"error-heading": errorHeading,
			"msg":           errorMsg,
		}, "layouts/main",
	)
}

func RenderInternalServerErrorPage(ctx *fiber.Ctx, err error) error {
	return RenderErrorPage(
		ctx, fiber.StatusInternalServerError, model.InternalServerError(errorfmt.Error(err)).CombinedMessage(),
	)
}
