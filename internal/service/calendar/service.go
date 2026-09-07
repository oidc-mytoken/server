// Package calendar provides a service layer for calendar operations.
package calendar

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/userrepo"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo/calendarrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/actions"
	calpkg "github.com/oidc-mytoken/server/internal/endpoints/notification/calendar/pkg"
	notpkg "github.com/oidc-mytoken/server/internal/endpoints/notification/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	notifier "github.com/oidc-mytoken/server/internal/notifier/client"
	"github.com/oidc-mytoken/server/internal/notifier/server/mailing"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/mytokenutils"
)

// Service is the calendar service singleton.
var Service = &service{}

type service struct{}

func newCalendarICS(id, description, icsPath string) string {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetDescription(
		fmt.Sprintf(
			"This calendar contains events and reminders for expiring mytokens issued from '%s'\n\n%s",
			config.Get().IssuerURL, description,
		),
	)
	cal.SetUrl(icsPath)
	return cal.Serialize()
}

func checkCalendarAccess(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, calendarID string, mtID mtid.MTID,
) *model.Response {
	ok, err := calendarrepo.MTIsForSameUserAsCalendar(rlog, tx, calendarID, mtID)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if !ok {
		return model.NotFoundErrorResponse("calendar not found")
	}
	return nil
}

func eventForMytoken(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID, comment string, unsubscribeOption bool, calendarID string,
) (*ics.VEvent, *model.Response) {
	mt, err := tree.SingleTokenEntry(rlog, tx, id)
	if err != nil {
		return nil, model.ErrorToInternalServerErrorResponse(err)
	}
	if mt.ExpiresAt == 0 {
		return nil, model.BadRequestErrorResponse("cannot create an expiration event for non-expiring mytokens")
	}
	event := ics.NewEvent(id.Hash())
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
		return nil, model.ErrorToInternalServerErrorResponse(err)
	}
	description += fmt.Sprintf(
		"To re-create a mytoken with similiar properties follow this link:\n"+
			"%s\n", recreateURL,
	)
	if unsubscribeOption {
		unsubscribeURL, err := actions.CreateRemoveFromCalendar(rlog, tx, id, calendarID)
		if err != nil {
			return nil, model.ErrorToInternalServerErrorResponse(err)
		}
		description += fmt.Sprintf(
			"To remove this mytoken from this calendar follow this link:\n"+
				"%s\n", unsubscribeURL,
		)
	}
	event.SetURL(recreateURL)
	event.SetDescription(description)
	createAlarms(event, mt, 30, 14, 7, 3, 1, 0)
	return event, nil
}

func addEventToCalendarICS(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, icsContent string, id mtid.MTID, comment, calendarID string,
) (string, error) {
	cal, err := ics.ParseCalendar(strings.NewReader(icsContent))
	if err != nil {
		return "", errors.WithStack(err)
	}
	event, errRes := eventForMytoken(rlog, tx, id, comment, true, calendarID)
	if errRes != nil {
		return "", errors.New("failed to create event")
	}
	cal.AddVEvent(event)
	return cal.Serialize(), nil
}

func updateCalendarDescriptionICS(
	icsContent, description string, issuerURL string,
) (string, error) {
	cal, err := ics.ParseCalendar(strings.NewReader(icsContent))
	if err != nil {
		return "", errors.WithStack(err)
	}
	cal.SetDescription(
		fmt.Sprintf(
			"This calendar contains events and reminders for expiring mytokens issued from '%s'\n\n%s",
			issuerURL, description,
		),
	)
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

// List returns all calendars for the user.
func (s *service) List(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData,
) *model.Response {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return model.BadRequestErrorResponse("calendar notifications are disabled")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, clientMetaData, api.CapabilityNotifyAnyTokenRead,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			infos, err := calendarrepo.List(rlog, tx, mt.ID)
			if err != nil {
				return err
			}
			res = &model.Response{
				Status: http.StatusOK,
				Response: &calpkg.CalendarListResponse{
					CalendarListResponse: api.CalendarListResponse{Calendars: infos},
				},
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, res, mt, *clientMetaData,
				api.EventCalendarListed, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

// Create creates a new calendar.
func (s *service) Create(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, description string, tags []string,
) *model.Response {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return model.BadRequestErrorResponse("calendar notifications are disabled")
	}
	id := utils.RandASCIIString(32)
	icsPath := calpkg.GetICSPath(id)
	calendarInfo := api.NotificationCalendar{
		ID:          id,
		ICSPath:     icsPath,
		Description: description,
	}
	dbInfo := calendarrepo.CalendarInfo{
		ID:          id,
		Description: db.NewNullString(description),
		ICS:         newCalendarICS(id, description, icsPath),
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, clientMetaData, api.CapabilityNotifyAnyToken,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if err := calendarrepo.Insert(rlog, tx, mt.ID, dbInfo); err != nil {
				return err
			}
			if len(tags) > 0 {
				if err := calendarrepo.LinkTags(rlog, tx, id, tags); err != nil {
					return err
				}
			}
			res = &model.Response{
				Status: http.StatusCreated,
				Response: &calpkg.CreateCalendarResponse{
					NotificationCalendar: calendarInfo,
				},
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, res, mt, *clientMetaData,
				api.EventCalendarCreated, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

// CheckAccess verifies the mytoken is not revoked and has access to the calendar.
func (s *service) CheckAccess(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData, calendarID string,
) *model.Response {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return model.BadRequestErrorResponse("calendar notifications are disabled")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if errRes := checkCalendarAccess(rlog, tx, calendarID, mt.ID); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

// Get returns a single calendar by ID.
func (s *service) Get(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData, calendarID string,
) *model.Response {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return model.BadRequestErrorResponse("calendar notifications are disabled")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			var rollback bool
			res, rollback = s.getLogic(rlog, tx, calendarID, mt.ID)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

func (s *service) getLogic(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, calendarID string, mtID mtid.MTID,
) (*model.Response, bool) {
	if errRes := checkCalendarAccess(rlog, tx, calendarID, mtID); errRes != nil {
		return errRes, true
	}
	info, err := calendarrepo.GetByID(rlog, tx, calendarID)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err), true
	}
	resInfo, err := info.ToCalendarInfoResponse(rlog, tx)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err), true
	}
	return &model.Response{
		Status:   http.StatusOK,
		Response: resInfo,
	}, false
}

// Delete deletes a calendar.
func (s *service) Delete(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, calendarID string,
) *model.Response {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return model.BadRequestErrorResponse("calendar notifications are disabled")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, clientMetaData, api.CapabilityNotifyAnyToken,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if err := calendarrepo.Delete(rlog, tx, mt.ID, calendarID); err != nil {
				return err
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				api.EventCalendarDeleted, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

// AddMytoken adds a mytoken to a calendar.
func (s *service) AddMytoken(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, calendarID, comment string, momID mtid.MOMID,
) *model.Response {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return model.BadRequestErrorResponse("calendar notifications are disabled")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if errRes := checkCalendarAccess(rlog, tx, calendarID, mt.ID); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			id, momMode, errRes := auth.ValidateCapabilityWithMomMode(
				rlog, tx, api.CapabilityTokeninfoNotify, api.CapabilityNotifyAnyToken,
				mt, momID, clientMetaData,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, tx, mt, clientMetaData)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			info, err := calendarrepo.GetByID(rlog, tx, calendarID)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if err = calendarrepo.AddMytokenToCalendar(rlog, tx, id, info.ID); err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			info.ICS, err = addEventToCalendarICS(rlog, tx, info.ICS, id, comment, calendarID)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if err = calendarrepo.UpdateICS(rlog, tx, id, calendarID, info.ICS); err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			event := api.EventNotificationSubscribed
			if momMode {
				event = api.EventNotificationSubscribedOther
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
				rlog, tx, res, mt, *clientMetaData,
				event, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

// RemoveMytoken removes a mytoken from a calendar.
func (s *service) RemoveMytoken(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, calendarID string, momID mtid.MOMID,
) *model.Response {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return model.BadRequestErrorResponse("calendar notifications are disabled")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if errRes := checkCalendarAccess(rlog, tx, calendarID, mt.ID); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			id, momMode, errRes := auth.ValidateCapabilityWithMomMode(
				rlog, tx, api.CapabilityTokeninfoNotify, api.CapabilityNotifyAnyToken,
				mt, momID, clientMetaData,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, tx, mt, clientMetaData)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if err := calendarrepo.RemoveMytokenFromCalendar(rlog, tx, id, calendarID); err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			// Update ICS to remove the event
			info, err := calendarrepo.GetByID(rlog, tx, calendarID)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			newICS := removeEventFromCalendarICS(info.ICS, id)
			if newICS != info.ICS {
				if err = calendarrepo.UpdateICS(rlog, tx, id, calendarID, newICS); err != nil {
					res = model.ErrorToInternalServerErrorResponse(err)
					return err
				}
			}
			event := api.EventNotificationUnsubscribed
			if momMode {
				event = api.EventNotificationUnsubscribedOther
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				event, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

// AddTag adds a tag to a calendar.
func (s *service) AddTag(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, calendarID, tag string,
) *model.Response {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return model.BadRequestErrorResponse("calendar notifications are disabled")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, clientMetaData, api.CapabilityNotifyAnyToken,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if errRes := checkCalendarAccess(rlog, tx, calendarID, mt.ID); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if err := calendarrepo.AddTag(rlog, tx, calendarID, tag); err != nil {
				return err
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				api.EventCalendarListed, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

// RemoveTag removes a tag from a calendar.
func (s *service) RemoveTag(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, calendarID, tag string,
) *model.Response {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return model.BadRequestErrorResponse("calendar notifications are disabled")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, clientMetaData, api.CapabilityNotifyAnyToken,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if errRes := checkCalendarAccess(rlog, tx, calendarID, mt.ID); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if err := calendarrepo.RemoveTag(rlog, tx, calendarID, tag); err != nil {
				return err
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				api.EventCalendarListed, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

// Update updates a calendar's description and/or tags.
func (s *service) Update(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, calendarID string,
	description *string, tags []string,
) *model.Response {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return model.BadRequestErrorResponse("calendar notifications are disabled")
	}
	if description == nil && tags == nil {
		return model.BadRequestErrorResponse("no update parameters provided")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, clientMetaData, api.CapabilityNotifyAnyToken,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if errRes := checkCalendarAccess(rlog, tx, calendarID, mt.ID); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			info, err := calendarrepo.GetByID(rlog, tx, calendarID)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			tagsUpdated := tags != nil
			descriptionUpdated := description != nil
			if descriptionUpdated {
				if err = calendarrepo.UpdateDescription(
					rlog, tx, mt.ID, calendarID, *description,
				); err != nil {
					return err
				}
				newICS, icsErr := updateCalendarDescriptionICS(info.ICS, *description, config.Get().IssuerURL)
				if icsErr != nil {
					rlog.WithError(icsErr).Warning(
						"failed to update ICS description, DB description already updated",
					)
				} else {
					if err = calendarrepo.UpdateICS(rlog, tx, mt.ID, calendarID, newICS); err != nil {
						res = model.ErrorToInternalServerErrorResponse(err)
						return err
					}
					info.ICS = newICS
				}
			}
			if tagsUpdated {
				if err = calendarrepo.LinkTags(rlog, tx, calendarID, tags); err != nil {
					return err
				}
			}
			event, eventComment := determineUpdateEvent(tagsUpdated, descriptionUpdated)
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
				rlog, tx, res, mt, *clientMetaData,
				event, eventComment, usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

func determineUpdateEvent(tagsUpdated, descriptionUpdated bool) (api.Event, string) {
	if tagsUpdated && descriptionUpdated {
		return api.EventCalendarUpdated, "updated tags and description"
	}
	if tagsUpdated {
		return api.EventCalendarTagsUpdated, ""
	}
	return api.EventCalendarUpdated, "updated description"
}

// removeEventFromCalendarICS removes an event for the given mtID from the ICS content.
func removeEventFromCalendarICS(icsContent string, mtID mtid.MTID) string {
	cal, err := ics.ParseCalendar(strings.NewReader(icsContent))
	if err != nil {
		return icsContent // return as-is if parsing fails
	}
	cal.RemoveEvent(mtID.Hash())
	cal.SetLastModified(time.Now())
	return cal.Serialize()
}

// removeStaleEvents removes events from the calendar that are no longer in mtids.
func removeStaleEvents(cal *ics.Calendar, mtids []string) map[string]struct{} {
	existing := make(map[string]struct{})
	for _, e := range cal.Events() {
		id := e.Id()
		if !utils.StringInSlice(id, mtids) {
			cal.RemoveEvent(id)
			cal.SetLastModified(time.Now())
		} else {
			existing[id] = struct{}{}
		}
	}
	return existing
}

// addMissingEvents adds events for mytokens that are not yet in the calendar.
func addMissingEvents(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, cal *ics.Calendar,
	mtids []string, existing map[string]struct{}, calendarID string,
) *model.Response {
	for _, mtidStr := range mtids {
		if _, ok := existing[mtidStr]; ok {
			continue
		}
		event, errRes := eventForMytoken(rlog, tx, mtid.FromHash(mtidStr), "", false, calendarID)
		if errRes != nil {
			return errRes
		}
		cal.AddVEvent(event)
		cal.SetLastModified(time.Now())
	}
	return nil
}

// syncICS synchronizes the calendar ICS with the current set of mytokens.
func syncICS(rlog log.Ext1FieldLogger, tx *sqlx.Tx, info calendarrepo.CalendarInfo) (string, *model.Response) {
	cal, err := ics.ParseCalendar(strings.NewReader(info.ICS))
	if err != nil {
		return "", model.ErrorToInternalServerErrorResponse(err)
	}
	mtids, err := calendarrepo.GetMTsInCalendar(rlog, tx, info.ID)
	if err != nil {
		return "", model.ErrorToInternalServerErrorResponse(err)
	}

	existing := removeStaleEvents(cal, mtids)
	if errRes := addMissingEvents(rlog, tx, cal, mtids, existing, info.ID); errRes != nil {
		return "", errRes
	}

	newICS := cal.Serialize()
	if newICS != info.ICS {
		if err = calendarrepo.UpdateICSInternal(rlog, tx, info.ID, newICS); err != nil {
			return "", model.ErrorToInternalServerErrorResponse(err)
		}
	}
	return newICS, nil
}

// GetICS returns the ICS content and tags for a calendar. This is used by the public ICS download endpoint.
func (s *service) GetICS(
	rlog log.Ext1FieldLogger, calendarID string,
) (icsContent string, tags []api.TagInfo, errRes *model.Response) {
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			var err error
			var info calendarrepo.CalendarInfo
			info, err = calendarrepo.GetByID(rlog, tx, calendarID)
			if err != nil {
				_, e := db.ParseError(err)
				if e != nil {
					errRes = model.ErrorToInternalServerErrorResponse(err)
				} else {
					errRes = model.NotFoundErrorResponse("calendar not found")
				}
				return err
			}
			tags, err = calendarrepo.GetCalendarTags(rlog, tx, calendarID)
			if err != nil {
				errRes = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			icsContent, errRes = syncICS(rlog, tx, info)
			if errRes != nil {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil {
		if errRes != nil {
			return
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		errRes = model.ErrorToInternalServerErrorResponse(err)
		return
	}
	return
}

func mailCalendarForMytoken(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID, name, comment, to string,
) (string, *model.Response) {
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

// CalendarEntryViaMail creates a calendar entry for a mytoken and sends it via email.
func (s *service) CalendarEntryViaMail(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, req notpkg.SubscribeNotificationRequest,
) *model.Response {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return model.BadRequestErrorResponse("calendar notifications are disabled")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			id := mt.ID
			momMode := req.MomID.Hash() != id.Hash()
			if momMode {
				id = req.MomID.MTID
				if res = auth.RequireMytokenIsParentOrCapability(
					rlog, tx, api.CapabilityTokeninfoNotify,
					api.CapabilityNotifyAnyToken, mt, id, clientMetaData,
				); res != nil {
					return errors.New("rollback")
				}
				if res = auth.RequireMytokensForSameUser(rlog, tx, id, mt.ID); res != nil {
					return errors.New("rollback")
				}
			}
			usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, tx, mt, clientMetaData)
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
				}, mt, *clientMetaData, mytokenEvent, "email calendar entry", usedRestriction,
				umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}
