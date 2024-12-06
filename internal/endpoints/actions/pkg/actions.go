package pkg

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/utils/utils"
)

// Actions
const (
	ActionRecreate             = "recreate_token"
	ActionVerifyEmail          = "verify_email"
	ActionRemoveFromCalendar   = "remove_from_calendar"
	ActionUnsubscribeScheduled = "unsubscribe_scheduled"
)

// CodeLifetimes holds the default lifetime of the different action codes
var CodeLifetimes = map[string]int{
	ActionVerifyEmail:          3600,
	ActionRecreate:             0,
	ActionRemoveFromCalendar:   0,
	ActionUnsubscribeScheduled: 0,
}

// ActionInfo is type for associating an Action with a Code
type ActionInfo struct {
	Action string
	Code   string
}

// CtxGetActionInfo obtains the ActionInfo from a fiber.Ctx
func CtxGetActionInfo(ctx *fiber.Ctx) ActionInfo {
	return ActionInfo{
		Action: ctx.Query("action"),
		Code:   ctx.Query("code"),
	}
}

// NewCode creates a new code
func NewCode() string {
	return utils.RandASCIIString(32)
}
