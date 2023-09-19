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
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo/calendarrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/notification/calendar/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	eventpkg "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/server/routes"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

//TODO events from eventservice
//TODO rotation

//TODO not found errors

func HandleGetICS(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle get ics calendar request")
	cid := ctx.Params("id")
	info, err := calendarrepo.GetByID(rlog, nil, cid)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	ctx.Set(fiber.HeaderContentType, "text/calendar")
	ctx.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, info.Name))
	return ctx.SendString(info.ICS)
}

func HandleAdd(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle add calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	var calendarInfo api.NotificationCalendar
	if err := errors.WithStack(ctx.BodyParser(&calendarInfo)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
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
	dbInfo := calendarrepo.CalendarInfo{
		ID:      id,
		Name:    calendarInfo.Name,
		ICSPath: icsPath,
		ICS:     cal.Serialize(),
	}
	if err := calendarrepo.Insert(rlog, nil, mt.ID, dbInfo); err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return model.Response{
		Status:   http.StatusCreated,
		Response: dbInfo,
	}.Send(ctx)
}

func HandleDelete(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	name := ctx.Params("name")
	rlog.WithField("calendar", name).Debug("Handle delete calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	if err := calendarrepo.Delete(rlog, nil, mt.ID, name); err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return model.Response{
		Status: http.StatusNoContent,
	}.Send(ctx)
}

func HandleGet(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle get calendar request")
	calendarName := ctx.Params("name")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	info, err := calendarrepo.Get(rlog, nil, mt.ID, calendarName)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return ctx.Redirect(info.ICSPath)
}

func HandleList(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle list calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	infos, err := calendarrepo.List(rlog, nil, mt.ID)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return model.Response{
		Status:   http.StatusOK,
		Response: infos,
	}.Send(ctx)
}

func HandleAddMytoken(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle add mytoken to calendar request")

	clientMetadata := ctxutils.ClientMetaData(ctx)
	calendarName := ctx.Params("name")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}

	var req pkg.AddMytokenToCalendarRequest
	fmt.Println(string(ctx.Body()))
	if err := errors.WithStack(ctx.BodyParser(&req)); err != nil {
		fmt.Println(errorfmt.Full(err))
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}

	id := mt.ID
	momMode := req.MomID.Hash() != id.Hash()
	if momMode {
		id = req.MomID.MTID
		if errRes = auth.RequireMytokenIsParentOrCapability(
			rlog, nil, api.CapabilityTokeninfoNotify,
			api.CapabilityNotifyAnyToken, mt, id,
		); errRes != nil {
			return errRes.Send(ctx)
		}
		if errRes = auth.RequireMytokensForSameUser(rlog, nil, id, mt.ID); errRes != nil {
			return errRes.Send(ctx)
		}
	}
	usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, nil, mt, clientMetadata.IP)
	if errRes != nil {
		return errRes.Send(ctx)
	}

	var res *model.Response
	_ = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			info, err := calendarrepo.Get(rlog, tx, id, calendarName)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			cal, err := ics.ParseCalendar(strings.NewReader(info.ICS))
			if err != nil {
				err = errors.WithStack(err)
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			event, err := eventForMytoken(rlog, tx, id, req.Comment)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			cal.AddVEvent(event)
			info.ICS = cal.Serialize()
			if err = calendarrepo.Update(rlog, tx, id, info); err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}

			res = &model.Response{
				Status:   http.StatusOK,
				Response: info,
			}

			if err = usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			mytokenEvent := eventpkg.FromNumber(
				eventpkg.NotificationSubscribed, fmt.Sprintf("calendar '%s'", info.Name),
			)
			if momMode {
				mytokenEvent.Type = eventpkg.NotificationSubscribedOther
			}
			if err = eventService.LogEvent(
				rlog, tx, eventService.MTEvent{
					Event: mytokenEvent,
					MTID:  mt.ID,
				}, *clientMetadata,
			); err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			return nil
		},
	)
	return res.Send(ctx)
}

func eventForMytoken(rlog logrus.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID, comment string) (*ics.VEvent, error) {
	var event *ics.VEvent
	if err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			mt, err := tree.SingleTokenEntry(rlog, tx, id)
			if err != nil {
				return err
			}
			if mt.ExpiresAt == 0 {
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
			if comment != "" {
				event.SetDescription(comment)
			}
			event.SetURL("https://mytok.eu/actions?action=recreate&code=foobar") //TODO
			createAlarms(event, mt, 30, 14, 7, 3, 1, 0)
			return nil
		},
	); err != nil {
		return nil, err
	}
	return event, nil
}

func createAlarms(event *ics.VEvent, info tree.MytokenEntry, triggerDaysBeforeExpiration ...int) {
	for _, d := range triggerDaysBeforeExpiration {
		if a := createAlarm(d, info); a != nil {
			event.Components = append(event.Components, a)
		}
	}
	return
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
