package notification

import (
	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/endpoints/notification/calendar"
)

// HandlePost is the main entry function for handling notification creation requests
func HandlePost(ctx *fiber.Ctx) error {
	//TODO switch
	return calendar.HandleCalendarEntryViaMail(ctx)
}
