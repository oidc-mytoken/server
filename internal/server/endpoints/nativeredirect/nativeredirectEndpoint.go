package nativeredirect

import (
	"github.com/gofiber/fiber/v2"

	"github.com/zachmann/mytoken/internal/server/db/dbrepo/supertokenrepo/transfercoderepo"
	"github.com/zachmann/mytoken/internal/server/model"
)

func HandleNativeRedirect(ctx *fiber.Ctx) error {
	poll := ctx.Params("poll")
	status, err := transfercoderepo.CheckTransferCode(nil, poll)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	if !status.Found || status.Expired || !status.RedirectURL.Valid {
		return model.Response{
			Status:   fiber.StatusNotFound,
			Response: "Polling Code not found",
		}.Send(ctx)
	}
	return ctx.Redirect(status.RedirectURL.String, fiber.StatusSeeOther)
}
