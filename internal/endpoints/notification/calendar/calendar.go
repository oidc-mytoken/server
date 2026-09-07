package calendar

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/endpoints/notification/calendar/pkg"
	notpkg "github.com/oidc-mytoken/server/internal/endpoints/notification/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/service/calendar"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleGetICS returns a calendar ics by its id (public, no auth required)
func HandleGetICS(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle get ics calendar request")
	cid := ctx.Params("id")
	icsContent, tags, errRes := calendar.Service.GetICS(rlog, cid)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	ctx.Set(fiber.HeaderContentType, "text/calendar")
	ctx.Set(fiber.HeaderContentDisposition, `attachment; filename=mytokens.ics`)
	// Include tags as JSON in a custom header for the calendar view
	if len(tags) > 0 {
		tagsJSON, _ := json.Marshal(tags)
		ctx.Set("X-Calendar-Tags", string(tagsJSON))
	}
	return ctx.SendString(icsContent)
}

// HandleAdd handles a request to create a new calendar
func HandleAdd(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle add calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &umt, ctx)
	if errRes != nil {
		return errRes
	}
	type createCalendarRequest struct {
		api.CreateCalendarRequest
		Tags []string `json:"tags"`
	}
	var request createCalendarRequest
	if err := errors.WithStack(ctx.BodyParser(&request)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	return calendar.Service.Create(
		rlog, mt, umt, ctxutils.ClientMetaData(ctx), request.Description, request.Tags,
	)
}

// HandleDelete deletes a calendar
func HandleDelete(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	calendarID := ctxutils.Params(ctx, "id")
	rlog.WithField("calendar", calendarID).Debug("Handle delete calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &umt, ctx)
	if errRes != nil {
		return errRes
	}
	res := calendar.Service.Delete(rlog, mt, umt, ctxutils.ClientMetaData(ctx), calendarID)
	if res == nil {
		return &model.Response{Status: http.StatusNoContent}
	}
	return res
}

// HandleGet looks up the id for a calendar name for the given user (by mytoken) and redirects to the ics endpoint
func HandleGet(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle get calendar request")
	calendarID := ctxutils.Params(ctx, "id")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &umt, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	errRes = calendar.Service.CheckAccess(rlog, mt, ctxutils.ClientMetaData(ctx), calendarID)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	return ctx.Redirect(pkg.GetICSPath(calendarID))
}

// HandleList lists all calendars for a user
func HandleList(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle list calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &umt, ctx)
	if errRes != nil {
		return errRes
	}
	return calendar.Service.List(rlog, mt, umt, ctxutils.ClientMetaData(ctx))
}

// HandleUpdate updates a calendar's description and/or tags
func HandleUpdate(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	calendarID := ctxutils.Params(ctx, "id")
	rlog.WithField("calendar", calendarID).Debug("Handle update calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &umt, ctx)
	if errRes != nil {
		return errRes
	}
	type updateCalendarRequest struct {
		Description *string  `json:"description"`
		Tags        []string `json:"tags"`
	}
	var req updateCalendarRequest
	if err := errors.WithStack(ctx.BodyParser(&req)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	return calendar.Service.Update(
		rlog, mt, umt, ctxutils.ClientMetaData(ctx), calendarID, req.Description, req.Tags,
	)
}

// HandleCalendarEntryViaMail creates a calendar entry for a mytoken and sends it via mail
func HandleCalendarEntryViaMail(
	ctx *fiber.Ctx, rlog logrus.Ext1FieldLogger, mt *mytoken.Mytoken,
	req notpkg.SubscribeNotificationRequest,
) *model.Response {
	rlog.Debug("Handle calendar entry via mail request")
	var umt universalmytoken.UniversalMytoken
	if mt != nil {
		umt = mt.ToUniversalMytoken()
	}
	return calendar.Service.CalendarEntryViaMail(rlog, mt, umt, ctxutils.ClientMetaData(ctx), req)
}

// HandleAddTag adds a tag to a calendar
func HandleAddTag(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	calendarID := ctxutils.Params(ctx, "id")
	rlog.WithField("calendar", calendarID).Debug("Handle add tag to calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &umt, ctx)
	if errRes != nil {
		return errRes
	}
	type addTagRequest struct {
		Tag string `json:"tag"`
	}
	var req addTagRequest
	if err := errors.WithStack(ctx.BodyParser(&req)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	if req.Tag == "" {
		return model.BadRequestErrorResponse("tag must not be empty")
	}
	res := calendar.Service.AddTag(rlog, mt, umt, ctxutils.ClientMetaData(ctx), calendarID, req.Tag)
	if res == nil {
		return &model.Response{Status: http.StatusNoContent}
	}
	return res
}

// HandleRemoveTag removes a tag from a calendar
func HandleRemoveTag(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	calendarID := ctxutils.Params(ctx, "id")
	rlog.WithField("calendar", calendarID).Debug("Handle remove tag from calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &umt, ctx)
	if errRes != nil {
		return errRes
	}
	type removeTagRequest struct {
		Tag string `json:"tag"`
	}
	var req removeTagRequest
	if err := errors.WithStack(ctx.BodyParser(&req)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	if req.Tag == "" {
		return model.BadRequestErrorResponse("tag must not be empty")
	}
	res := calendar.Service.RemoveTag(rlog, mt, umt, ctxutils.ClientMetaData(ctx), calendarID, req.Tag)
	if res == nil {
		return &model.Response{Status: http.StatusNoContent}
	}
	return res
}

// HandleAddMytoken handles a request to add a mytoken to a calendar
func HandleAddMytoken(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle add mytoken to calendar request")

	calendarID := ctxutils.Params(ctx, "id")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &umt, ctx)
	if errRes != nil {
		return errRes
	}
	var req pkg.AddMytokenToCalendarRequest
	if err := errors.WithStack(ctx.BodyParser(&req)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	return calendar.Service.AddMytoken(
		rlog, mt, umt, ctxutils.ClientMetaData(ctx), calendarID, req.Comment, req.MomID,
	)
}

// HandleRemoveMytoken handles a request to remove a mytoken from a calendar
func HandleRemoveMytoken(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle remove mytoken from calendar request")

	calendarID := ctxutils.Params(ctx, "id")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &umt, ctx)
	if errRes != nil {
		return errRes
	}
	type removeMytokenRequest struct {
		MomID mtid.MOMID `json:"mom_id"`
	}
	var req removeMytokenRequest
	if err := errors.WithStack(ctx.BodyParser(&req)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	momID := mt.ID.MomID()
	if req.MomID.Hash() != "" {
		momID = req.MomID
	}
	res := calendar.Service.RemoveMytoken(rlog, mt, umt, ctxutils.ClientMetaData(ctx), calendarID, momID)
	if res == nil {
		return &model.Response{Status: http.StatusNoContent}
	}
	return res
}
