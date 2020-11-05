package redirect

import (
	"log"

	"github.com/zachmann/mytoken/internal/model"

	"github.com/zachmann/mytoken/internal/oidc/authcode"

	"github.com/gofiber/fiber/v2"
)

func HandleOIDCRedirect(ctx *fiber.Ctx) error {
	log.Print("Handle redirect")
	error := ctx.Query("error")
	state := ctx.Query("state")
	log.Printf("error: '%s'", error)
	log.Printf("state: '%s'", state)
	if error != "" {
		if state != "" {
			//TODO delete AuthInfo (and pollingCode)
		}
		errorDescription := ctx.Query("error_description")
		errorRes := model.Response{
			Status:   fiber.StatusInternalServerError,
			Response: model.OIDCError(error, errorDescription),
		}
		return errorRes.Send(ctx)
	}
	code := ctx.Query("code")
	networkData := model.NetworkData{
		IP:        ctx.IP(),
		UserAgent: string(ctx.Request().Header.UserAgent()),
	}
	res := authcode.CodeExchange(state, code, networkData)
	return res.Send(ctx)
}
