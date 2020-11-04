package redirect

import (
	"fmt"

	"github.com/zachmann/mytoken/internal/oidc/authcode"

	"github.com/gofiber/fiber/v2"
)

func HandleOIDCRedirect(ctx *fiber.Ctx) error {
	error := ctx.Params("error")
	state := ctx.Params("state")
	if error != "" {
		errorDescription := ctx.Params("error_description")
		if errorDescription != "" {
			error = fmt.Sprintf("%s: %s", error, errorDescription)
		}
		if state != "" {
			//TODO delete AuthInfo (and pollingCode)
		}
		ctx.SendStatus(fiber.StatusBadRequest)
		return ctx.SendString(fmt.Sprintf("error: %s", error))
	}
	code := ctx.Params("code")
	res := authcode.CodeExchange(state, code, ctx.IP())
	return res.Send(ctx)
}
