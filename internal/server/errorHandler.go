package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	fiberUtils "github.com/gofiber/fiber/v2/utils"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/server/internal/model"
)

func handleError(ctx *fiber.Ctx, err error) error {
	// Statuscode defaults to 500
	code := fiber.StatusInternalServerError
	msg := err.Error()

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		msg = e.Message
	}

	if ctx.Accepts(fiber.MIMETextHTML, fiber.MIMETextHTMLCharsetUTF8) != "" {
		return handleErrorHTML(ctx, code, msg)
	}
	return handleErrorJSON(ctx, code, msg)

}

func handleErrorHTML(ctx *fiber.Ctx, code int, msg string) error {
	var err error
	errorTemplateData := map[string]interface{}{
		"empty-navbar": true,
		"msg":          msg,
	}
	switch code {
	case fiber.StatusNotFound,
		fiber.StatusMethodNotAllowed,
		fiber.StatusTooManyRequests,
		fiber.StatusInternalServerError,
		fiber.StatusNotImplemented,
		fiber.StatusHTTPVersionNotSupported:
		ctx.Status(code)
		err = ctx.Render(fmt.Sprintf("sites/%d", code), errorTemplateData, "layouts/main")
	default:
		return handleErrorJSON(ctx, code, msg)
	}
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return nil
}

func handleErrorJSON(ctx *fiber.Ctx, code int, msg string) error {
	return model.Response{
		Status: code,
		Response: api.Error{
			Error:            fiberUtils.StatusMessage(code),
			ErrorDescription: msg,
		},
	}.Send(ctx)
}
