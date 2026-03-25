package ctxutils

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/server/spa"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

// RenderErrorPage renders an error page
func RenderErrorPage(ctx *fiber.Ctx, status int, errorMsg string, optionalErrorHeading ...string) error {
	var errorHeading string
	if len(optionalErrorHeading) > 0 {
		errorHeading = optionalErrorHeading[0]
	}
	return RenderExtendedErrorPage(ctx, status, errorMsg, errorHeading, "")
}

// RenderExtendedErrorPage renders an error page with additional html content
func RenderExtendedErrorPage(
	ctx *fiber.Ctx, status int, errorMsg,
	optionalErrorHeading, additionalHTML string,
) error {
	errorHeading := http.StatusText(status)
	if optionalErrorHeading != "" {
		errorHeading = optionalErrorHeading
	}
	return ctx.Status(status).Render(
		"sites/error", map[string]interface{}{
			"empty-navbar":    true,
			"error-heading":   errorHeading,
			"msg":             errorMsg,
			"additional-html": additionalHTML,
		}, "layouts/main",
	)
}

// RenderInternalServerErrorPage renders an error page for a passed error as an internal server error
func RenderInternalServerErrorPage(ctx *fiber.Ctx, err error) error {
	return RenderErrorPage(
		ctx, fiber.StatusInternalServerError, model.InternalServerError(errorfmt.Error(err)).CombinedMessage(),
	)
}

// RenderActionResultPage renders an action result page (for email verification, etc.)
// Uses standalone HTML when SPA is active, falls back to Mustache templates otherwise
func RenderActionResultPage(ctx *fiber.Ctx, status int, title, message string, isSuccess bool) error {
	if spa.Available && config.Get().Features.WebInterface.UseSPA {
		return spa.RenderActionResultPage(ctx, status, title, message, isSuccess)
	}
	// Fall back to existing Mustache rendering
	return RenderErrorPage(ctx, status, message, title)
}
