package server

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	fiberUtils "github.com/gofiber/fiber/v2/utils"
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/server/apipath"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

func handleError(ctx *fiber.Ctx, err error) error {
	// Status code defaults to 500
	code := fiber.StatusInternalServerError
	msg := errorfmt.Error(err)
	rlog := logger.GetRequestLogger(ctx)

	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
		msg = e.Error()
	}
	if code >= 500 {
		rlog.Errorf("%s", errorfmt.Full(err))
	}

	if ctx.Accepts(fiber.MIMETextHTML, fiber.MIMETextHTMLCharsetUTF8) != "" && !strings.HasPrefix(
		ctx.Path(), apipath.Prefix,
	) {
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
		log := logger.GetRequestLogger(ctx)
		log.WithError(err).Error()
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
