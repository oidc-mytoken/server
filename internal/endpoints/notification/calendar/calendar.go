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
	pkg2 "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mailing"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	eventpkg "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/rotation"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/server/routes"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

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
		return model.NotFoundErrorResponse("calendar not found").Send(ctx)
	}
	ctx.Set(fiber.HeaderContentType, "text/calendar")
	ctx.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, info.Name))
	return ctx.SendString(info.ICS)
}

// HandleAdd handles a request to create a new calendar
func HandleAdd(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle add calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt,
		ctxutils.ClientMetaData(ctx).IP, api.CapabilityNotifyAnyToken,
	)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	var calendarInfo api.NotificationCalendar
	if err := errors.WithStack(ctx.BodyParser(&calendarInfo)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	if calendarInfo.Name == "" {
		return model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError("required parameter 'name' is missing"),
		}.Send(ctx)
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
	res := model.Response{
		Status:   http.StatusCreated,
		Response: pkg.CreateCalendarResponse{CalendarInfo: dbInfo},
	}
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err := calendarrepo.Insert(rlog, tx, mt.ID, dbInfo); err != nil {
				return err
			}
			tokenUpdate, err := rotation.RotateMytokenAfterOtherForResponse(
				rlog, tx, umt.JWT, mt, *ctxutils.ClientMetaData(ctx), umt.OriginalTokenType,
			)
			if err != nil {
				return err
			}
			if tokenUpdate != nil {
				res.Cookies = []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)}
				resData := res.Response.(pkg.CreateCalendarResponse)
				resData.TokenUpdate = tokenUpdate
				res.Response = resData
			}
			if err = usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
				return err
			}
			return eventService.LogEvent(
				rlog, tx, eventService.MTEvent{
					Event: eventpkg.FromNumber(eventpkg.CalendarCreated, calendarInfo.Name),
					MTID:  mt.ID,
				},
				*ctxutils.ClientMetaData(ctx),
			)
		},
	); err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return res.Send(ctx)
}

// HandleDelete deletes a calendar
func HandleDelete(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	name := ctx.Params("name")
	rlog.WithField("calendar", name).Debug("Handle delete calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt,
		ctxutils.ClientMetaData(ctx).IP, api.CapabilityNotifyAnyToken,
	)
	if errRes != nil {
		return errRes.Send(ctx)
	}

	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err := calendarrepo.Delete(rlog, tx, mt.ID, name); err != nil {
				return err
			}
			tokenUpdate, err := rotation.RotateMytokenAfterOtherForResponse(
				rlog, tx, umt.JWT, mt, *ctxutils.ClientMetaData(ctx), umt.OriginalTokenType,
			)
			if err != nil {
				return err
			}
			if tokenUpdate != nil {
				res = &model.Response{
					Status: fiber.StatusOK,
					Response: pkg2.OnlyTokenUpdateRes{
						TokenUpdate: tokenUpdate,
					},
					Cookies: []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)},
				}
			}
			if err = usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
				return err
			}
			return eventService.LogEvent(
				rlog, tx, eventService.MTEvent{
					Event: eventpkg.FromNumber(eventpkg.CalendarDeleted, name),
					MTID:  mt.ID,
				},
				*ctxutils.ClientMetaData(ctx),
			)
		},
	); err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	if res != nil {
		return res.Send(ctx)
	}
	return model.Response{
		Status: http.StatusNoContent,
	}.Send(ctx)
}

// HandleGet looks up the id for a calendar name for the given user (by mytoken) and redirects to the ics endpoint
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
		_, e := db.ParseError(err)
		if e != nil {
			return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
		}
		return model.NotFoundErrorResponse("calendar not found").Send(ctx)
	}
	return ctx.Redirect(info.ICSPath)
}

// HandleList lists all calendars for a user
func HandleList(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle list calendar request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt,
		ctxutils.ClientMetaData(ctx).IP, api.CapabilityNotifyAnyTokenRead,
	)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	var res *model.Response
	_ = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			infos, err := calendarrepo.List(rlog, tx, mt.ID)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			resData := pkg.CalendarListResponse{Calendars: infos}
			res = &model.Response{
				Status:   fiber.StatusOK,
				Response: resData,
			}

			tokenUpdate, err := rotation.RotateMytokenAfterOtherForResponse(
				rlog, tx, umt.JWT, mt, *ctxutils.ClientMetaData(ctx), umt.OriginalTokenType,
			)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if tokenUpdate != nil {
				res.Cookies = []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)}
				resData.TokenUpdate = tokenUpdate
				res.Response = resData
			}
			if err = usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
				return err
			}
			return eventService.LogEvent(
				rlog, tx, eventService.MTEvent{
					Event: eventpkg.FromNumber(eventpkg.CalendarListed, ""),
					MTID:  mt.ID,
				},
				*ctxutils.ClientMetaData(ctx),
			)
		},
	)
	return res.Send(ctx)
}

// HandleCalendarEntryViaMail creates a calendar entry for a mytoken and sends it via mail
func HandleCalendarEntryViaMail(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle calendar entry via mail request")

	clientMetadata := ctxutils.ClientMetaData(ctx)
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}

	var req pkg.AddMytokenToCalendarRequest
	if err := errors.WithStack(ctx.BodyParser(&req)); err != nil {
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
			mailInfo, err := userrepo.GetMail(rlog, tx, id)
			found, err := db.ParseError(err)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if !found {
				res = &model.Response{
					Status:   http.StatusPreconditionRequired,
					Response: api.ErrorMailRequired,
				}
				return errors.New("dummy")
			}
			if !mailInfo.MailVerified {
				res = &model.Response{
					Status:   http.StatusPreconditionRequired,
					Response: api.ErrorMailNotVerified,
				}
				return errors.New("dummy")
			}
			mtInfo, err := tree.SingleTokenEntry(rlog, tx, id)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			calText, err := mailCalendarForMytoken(rlog, tx, id, mtInfo.Name.String, req.Comment, mailInfo.Mail)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}

			filename := mtInfo.Name.String
			if filename == "" {
				filename = id.Hash()
			}
			err = mailing.ICSMailSender.Send(
				mailInfo.Mail,
				fmt.Sprintf("Mytoken Expiration Calendar Reminder for '%s'", filename),
				"You can add the event to your calendar to be notified before the mytoken expires.",
				mailing.Attachment{
					Reader:      strings.NewReader(calText),
					Filename:    filename + ".ics",
					ContentType: "text/calendar",
				},
			)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}

			if err = usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			mytokenEvent := eventpkg.FromNumber(eventpkg.NotificationSubscribed, "email calendar entry")
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
			res = &model.Response{
				Status: http.StatusNoContent,
			}
			tokenUpdate, err := rotation.RotateMytokenAfterOtherForResponse(
				rlog, tx, umt.JWT, mt, *ctxutils.ClientMetaData(ctx), umt.OriginalTokenType,
			)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if tokenUpdate != nil {
				res.Cookies = []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)}
				res.Response = pkg2.OnlyTokenUpdateRes{TokenUpdate: tokenUpdate}
			}
			return nil
		},
	)
	return res.Send(ctx)
}

// HandleAddMytoken handles a request to add a mytoken to a calendar
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
				_, e := db.ParseError(err)
				if e != nil {
					res = model.ErrorToInternalServerErrorResponse(err)
				}
				res = model.NotFoundErrorResponse("calendar not found")
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
			event, err := eventForMytoken(rlog, tx, id, req.Comment, true, calendarName)
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

			res = &model.Response{
				Status:   http.StatusOK,
				Response: info,
			}
			tokenUpdate, err := rotation.RotateMytokenAfterOtherForResponse(
				rlog, tx, umt.JWT, mt, *ctxutils.ClientMetaData(ctx), umt.OriginalTokenType,
			)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if tokenUpdate != nil {
				res.Cookies = []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)}
				res.Response = pkg2.OnlyTokenUpdateRes{TokenUpdate: tokenUpdate}
			}

			return nil
		},
	)
	return res.Send(ctx)
}

func eventForMytoken(
	rlog logrus.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID, comment string,
	unsubscribeOption bool, calendarName string,
) (*ics.VEvent, error) {
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
			description := comment
			if description != "" {
				description += "\n\n"
			}
			recreateURL, err := actions.CreateRecreateToken(rlog, tx, id)
			if err != nil {
				return err
			}
			description += fmt.Sprintf(
				"To re-create a mytoken with similiar properties follow this link:\n"+
					"%s\n", recreateURL,
			)
			if unsubscribeOption {
				unsubscribeURL, err := actions.CreateRemoveFromCalendar(rlog, tx, id, calendarName)
				if err != nil {
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
	); err != nil {
		return nil, err
	}
	return event, nil
}
func mailCalendarForMytoken(rlog logrus.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID, name, comment, to string) (
	string,
	error,
) {
	event, err := eventForMytoken(rlog, tx, id, comment, false, "")
	if err != nil {
		return "", err
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
