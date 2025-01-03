package calendar

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/userrepo"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo/calendarrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/actions"
	"github.com/oidc-mytoken/server/internal/endpoints/notification/calendar/pkg"
	pkg4 "github.com/oidc-mytoken/server/internal/endpoints/notification/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	notifier "github.com/oidc-mytoken/server/internal/notifier/client"
	"github.com/oidc-mytoken/server/internal/notifier/server/mailing"
	"github.com/oidc-mytoken/server/internal/server/routes"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/internal/utils/mytokenutils"
)

var calendarNotFoundError = model.NotFoundErrorResponse("calendar not found")

// HandleGetICS returns a calendar ics by its id
func HandleGetICS(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle get ics calendar request")
	cid := ctx.Params("id")
	var info calendarrepo.CalendarInfo
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			var err error
			info, err = calendarrepo.GetByID(rlog, tx, cid)
			if err != nil {
				return err
			}

			cal, err := ics.ParseCalendar(strings.NewReader(info.ICS))
			if err != nil {
				return err
			}
			mtids, err := calendarrepo.GetMTsInCalendar(rlog, tx, info.ID)
			if err != nil {
				return err
			}
			for _, e := range cal.Events() {
				id := e.Id()
				if !utils.StringInSlice(id, mtids) {
					cal.RemoveEvent(id)
					cal.SetLastModified(time.Now())
				}
			}
			newICS := cal.Serialize()
			if newICS != info.ICS {
				info.ICS = newICS
				return calendarrepo.UpdateInternal(rlog, tx, info)
			}
			return nil
		},
	); err != nil {
		_, e := db.ParseError(err)
		if e != nil {
			return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
		}
		return calendarNotFoundError.Send(ctx)
	}
	ctx.Set(fiber.HeaderContentType, "text/calendar")
	ctx.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename=%q`, info.Name))
	return ctx.SendString(info.ICS)
}

// HandleAdd handles a request to create a new calendar
func HandleAdd(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle add calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt,
		ctxutils.ClientMetaData(ctx), api.CapabilityNotifyAnyToken,
	)
	if errRes != nil {
		return errRes
	}
	var calendarInfo api.NotificationCalendar
	if err := errors.WithStack(ctx.BodyParser(&calendarInfo)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	if calendarInfo.Name == "" {
		return model.BadRequestErrorResponse("required parameter 'name' is missing")
	}

	id := utils.RandASCIIString(32)
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetName(calendarInfo.Name)
	cal.SetDescription(
		fmt.Sprintf(
			"This calendar contains events and reminders for expiring mytokens issued from '%s'",
			config.Get().IssuerURL,
		),
	)
	icsPath := utils.CombineURLPath(routes.CalendarDownloadEndpoint, id)
	cal.SetUrl(icsPath)
	calendarInfo.ICSPath = icsPath
	dbInfo := calendarrepo.CalendarInfo{
		ID:      id,
		Name:    calendarInfo.Name,
		ICSPath: icsPath,
		ICS:     cal.Serialize(),
	}
	res := &model.Response{
		Status:   http.StatusCreated,
		Response: &pkg.CreateCalendarResponse{NotificationCalendar: calendarInfo},
	}
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err := calendarrepo.Insert(rlog, tx, mt.ID, dbInfo); err != nil {
				return err
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, res, mt, *ctxutils.ClientMetaData(ctx),
				api.EventCalendarCreated, calendarInfo.Name, usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil && res == nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

// HandleDelete deletes a calendar
func HandleDelete(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	name := ctxutils.Params(ctx, "name")
	rlog.WithField("calendar", name).Debug("Handle delete calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt,
		ctxutils.ClientMetaData(ctx), api.CapabilityNotifyAnyToken,
	)
	if errRes != nil {
		return errRes
	}

	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err := calendarrepo.Delete(rlog, tx, mt.ID, name); err != nil {
				return err
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *ctxutils.ClientMetaData(ctx),
				api.EventCalendarDeleted, name, usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil && res == nil {
		res = model.ErrorToInternalServerErrorResponse(err)
	}
	if res == nil {
		res = &model.Response{
			Status: http.StatusNoContent,
		}
	}
	return res
}

// HandleGet looks up the id for a calendar name for the given user (by mytoken) and redirects to the ics endpoint
func HandleGet(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle get calendar request")
	calendarName := ctxutils.Params(ctx, "name")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	info, err := calendarrepo.Get(rlog, nil, mt.ID, calendarName)
	if err != nil {
		_, e := db.ParseError(err)
		if e != nil {
			return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
		}
		return calendarNotFoundError.Send(ctx)
	}
	return ctx.Redirect(info.ICSPath)
}

// HandleList lists all calendars for a user
func HandleList(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle list calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt,
		ctxutils.ClientMetaData(ctx), api.CapabilityNotifyAnyTokenRead,
	)
	if errRes != nil {
		return errRes
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			infos, err := calendarrepo.List(rlog, tx, mt.ID)
			if err != nil {
				return err
			}
			res = &model.Response{
				Status:   fiber.StatusOK,
				Response: &pkg.CalendarListResponse{CalendarListResponse: api.CalendarListResponse{Calendars: infos}},
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, res, mt, *ctxutils.ClientMetaData(ctx),
				api.EventCalendarListed, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil && res == nil {
		res = model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

// HandleCalendarEntryViaMail creates a calendar entry for a mytoken and sends it via mail
func HandleCalendarEntryViaMail(
	ctx *fiber.Ctx, rlog logrus.Ext1FieldLogger, mt *mytoken.Mytoken,
	req pkg4.SubscribeNotificationRequest,
) *model.Response {
	rlog.Debug("Handle calendar entry via mail request")
	var res *model.Response
	_ = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			clientMetadata := ctxutils.ClientMetaData(ctx)
			id := mt.ID
			momMode := req.MomID.Hash() != id.Hash()
			if momMode {
				id = req.MomID.MTID
				if res = auth.RequireMytokenIsParentOrCapability(
					rlog, tx, api.CapabilityTokeninfoNotify,
					api.CapabilityNotifyAnyToken, mt, id, clientMetadata,
				); res != nil {
					return errors.New("rollback")
				}
				if res = auth.RequireMytokensForSameUser(rlog, tx, id, mt.ID); res != nil {
					return errors.New("rollback")
				}
			}
			usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, tx, mt, clientMetadata)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			mailInfo, errRes, err := userrepo.GetAndCheckMail(rlog, tx, id)
			if err != nil {
				res = errRes
				return err
			}
			mtInfo, err := tree.SingleTokenEntry(rlog, tx, id)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			calText, errRes := mailCalendarForMytoken(
				rlog, tx, id, mtInfo.Name.String, req.Comment, mailInfo.Mail.String,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}

			filename := mtInfo.Name.String
			if filename == "" {
				filename = id.Hash()
			}
			notifier.SendICSMail(
				mailInfo.Mail.String,
				fmt.Sprintf("Mytoken Expiration Calendar Reminder for '%s'", filename),
				"You can add the event to your calendar to be notified before the mytoken expires.",
				mailing.Attachment{
					Reader:      strings.NewReader(calText),
					Filename:    filename + ".ics",
					ContentType: "text/calendar",
				},
			)

			mytokenEvent := api.EventNotificationSubscribed
			if momMode {
				mytokenEvent = api.EventNotificationSubscribedOther
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, &model.Response{
					Status: http.StatusNoContent,
				}, mt, *ctxutils.ClientMetaData(ctx), mytokenEvent, "email calendar entry", usedRestriction,
				req.Mytoken.JWT, req.Mytoken.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	)
	return res
}

// HandleAddMytoken handles a request to add a mytoken to a calendar
func HandleAddMytoken(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle add mytoken to calendar request")

	clientMetadata := ctxutils.ClientMetaData(ctx)
	calendarName := ctxutils.Params(ctx, "name")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes
	}

	var req pkg.AddMytokenToCalendarRequest
	if err := errors.WithStack(ctx.BodyParser(&req)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}

	id, momMode, errRes := validateMomMode(rlog, mt, req, clientMetadata)
	if errRes != nil {
		return errRes
	}
	usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, nil, mt, clientMetadata)
	if errRes != nil {
		return errRes
	}

	var res *model.Response
	_ = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			info, err := calendarrepo.Get(rlog, tx, id, calendarName)
			if err != nil {
				_, e := db.ParseError(err)
				if e == nil {
					res = calendarNotFoundError
				} else {
					res = model.ErrorToInternalServerErrorResponse(err)
				}
				return err
			}
			if err = calendarrepo.AddMytokenToCalendar(rlog, tx, id, info.ID); err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			cal, err := ics.ParseCalendar(strings.NewReader(info.ICS))
			if err != nil {
				err = errors.WithStack(err)
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			event, errRes := eventForMytoken(rlog, tx, id, req.Comment, true, calendarName)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			cal.AddVEvent(event)
			info.ICS = cal.Serialize()
			if err = calendarrepo.Update(rlog, tx, id, info); err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}

			mytokenEvent := api.EventNotificationSubscribed
			if momMode {
				mytokenEvent = api.EventNotificationSubscribedOther
			}
			resInfo, err := info.ToCalendarInfoResponse(rlog, tx)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			res = &model.Response{
				Status:   http.StatusOK,
				Response: resInfo,
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, res, mt, *ctxutils.ClientMetaData(ctx),
				mytokenEvent, fmt.Sprintf("calendar '%s'", info.Name), usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	)
	return res
}

func validateMomMode(
	rlog logrus.Ext1FieldLogger, mt *mytoken.Mytoken, req pkg.AddMytokenToCalendarRequest,
	clientMetadata *api.ClientMetaData,
) (mtid.MTID, bool, *model.Response) {
	id := mt.ID
	momMode := req.MomID.Hash() != id.Hash()
	if momMode {
		id = req.MomID.MTID
		if errRes := auth.RequireMytokenIsParentOrCapability(
			rlog, nil, api.CapabilityTokeninfoNotify, api.CapabilityNotifyAnyToken, mt, id, clientMetadata,
		); errRes != nil {
			return id, momMode, errRes
		}
		if errRes := auth.RequireMytokensForSameUser(rlog, nil, id, mt.ID); errRes != nil {
			return id, momMode, errRes
		}
	}
	return id, momMode, nil
}

func eventForMytoken(
	rlog logrus.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID, comment string,
	unsubscribeOption bool, calendarName string,
) (event *ics.VEvent, errRes *model.Response) {
	_ = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			mt, err := tree.SingleTokenEntry(rlog, tx, id)
			if err != nil {
				errRes = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if mt.ExpiresAt == 0 {
				errRes = model.BadRequestErrorResponse("cannot create an expiration event for non-expiring mytokens")
				return nil
			}
			event = ics.NewEvent(id.Hash())
			now := time.Now()
			event.SetCreatedTime(now)
			event.SetDtStampTime(now)
			event.SetModifiedAt(now)
			event.SetStartAt(mt.ExpiresAt.Time())
			event.SetEndAt(mt.ExpiresAt.Time())
			title := "Mytoken expires"
			if mt.Name.Valid {
				title = fmt.Sprintf("Mytoken '%s' expires", mt.Name.String)
			}
			event.SetSummary(title)
			description := comment
			if description != "" {
				description += "\n\n"
			}
			recreateURL, err := actions.CreateRecreateToken(rlog, tx, id)
			if err != nil {
				errRes = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			description += fmt.Sprintf(
				"To re-create a mytoken with similiar properties follow this link:\n"+
					"%s\n", recreateURL,
			)
			if unsubscribeOption {
				unsubscribeURL, err := actions.CreateRemoveFromCalendar(rlog, tx, id, calendarName)
				if err != nil {
					errRes = model.ErrorToInternalServerErrorResponse(err)
					return err
				}
				description += fmt.Sprintf(
					"To remove this mytoken from calendar '%s' follow this link:\n"+
						"%s\n", calendarName, unsubscribeURL,
				)
			}
			event.SetURL(recreateURL)
			event.SetDescription(description)
			createAlarms(event, mt, 30, 14, 7, 3, 1, 0)
			return nil
		},
	)
	return
}
func mailCalendarForMytoken(rlog logrus.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID, name, comment, to string) (
	string,
	*model.Response,
) {
	event, errRes := eventForMytoken(rlog, tx, id, comment, false, "")
	if errRes != nil {
		return "", errRes
	}
	event.AddAttendee(to)
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodRequest)
	cal.SetName(name)
	cal.AddVEvent(event)
	return cal.Serialize(), nil
}

func createAlarms(event *ics.VEvent, info tree.MytokenEntry, triggerDaysBeforeExpiration ...int) {
	for _, d := range triggerDaysBeforeExpiration {
		if a := createAlarm(d, info); a != nil {
			event.Components = append(event.Components, a)
		}
	}
}

func createAlarm(daysBeforeExpiration int, info tree.MytokenEntry) *ics.VAlarm {
	now := time.Now()
	expiresAt := info.ExpiresAt.Time()
	createdAt := info.CreatedAt.Time()
	triggerTime := expiresAt.Add(time.Duration(-24*daysBeforeExpiration) * time.Hour)
	if triggerTime.Before(now) {
		return nil
	}
	if triggerTime.Before(createdAt.Add(expiresAt.Sub(createdAt) / 2)) {
		return nil
	}
	alarm := &ics.VAlarm{
		ComponentBase: ics.ComponentBase{},
	}
	alarm.SetAction(ics.ActionDisplay)
	alarm.SetTrigger(fmt.Sprintf("-PT%dD", daysBeforeExpiration))
	return alarm
}
