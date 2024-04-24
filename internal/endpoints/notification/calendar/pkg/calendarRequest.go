package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

// AddMytokenToCalendarRequest is type holding the request to add a mytoken to a calendar
type AddMytokenToCalendarRequest struct {
	api.AddMytokenToCalendarRequest
	MomID mtid.MOMID `json:"mom_id" xml:"mom_id" form:"mom_id"`
}

// CreateCalendarResponse is the response returned when a new calendar is created
type CreateCalendarResponse struct {
	api.NotificationCalendar
	pkg.OnlyTokenUpdateRes
}

// CalendarListResponse is the response returned to list all calendars of a user
type CalendarListResponse struct {
	api.CalendarListResponse
	pkg.OnlyTokenUpdateRes
}

// CalendarInfoResponse is a type for holding a response with information about a single calendar
type CalendarInfoResponse struct {
	api.CalendarInfo
	pkg.OnlyTokenUpdateRes
}
