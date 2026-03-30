package ctxutils

import (
	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/server/spa"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

// RenderErrorPage renders an error page using the SPA
func RenderErrorPage(ctx *fiber.Ctx, status int, errorMsg string, _ ...string) error {
	// Let the SPA handle error display via client-side routing
	handler := spa.HandleSPAFallback()
	if handler != nil {
		ctx.Status(status)
		return handler(ctx)
	}
	// Fallback to JSON if SPA is not available
	return ctx.Status(status).JSON(
		fiber.Map{
			"error": errorMsg,
		},
	)
}

// RenderExtendedErrorPage renders an error page with additional html content
func RenderExtendedErrorPage(
	ctx *fiber.Ctx, status int, errorMsg,
	optionalErrorHeading, _ string,
) error {
	return RenderErrorPage(ctx, status, errorMsg, optionalErrorHeading)
}

// RenderInternalServerErrorPage renders an error page for a passed error as an internal server error
func RenderInternalServerErrorPage(ctx *fiber.Ctx, err error) error {
	return RenderErrorPage(
		ctx, fiber.StatusInternalServerError, model.InternalServerError(errorfmt.Error(err)).CombinedMessage(),
	)
}

// RenderActionResultPage renders an action result page (for email verification, etc.)
func RenderActionResultPage(ctx *fiber.Ctx, status int, title, message string, isSuccess bool) error {
	return spa.RenderActionResultPage(ctx, status, title, message, isSuccess)
}
