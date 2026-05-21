package ssh

import (
	"encoding/json"
	"net/http"

	"github.com/gliderlabs/ssh"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo/calendarrepo"
	calpkg "github.com/oidc-mytoken/server/internal/endpoints/notification/calendar/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

func requireCalendarAccess(rlog log.Ext1FieldLogger, s ssh.Session, calendarID string, mtID mtid.MTID) error {
	ok, err := calendarrepo.MTIsForSameUserAsCalendar(rlog, nil, calendarID, mtID)
	if err != nil {
		return writeErrRes(s, model.ErrorToInternalServerErrorResponse(err))
	}
	if !ok {
		return writeErrRes(s, model.NotFoundErrorResponse("calendar not found"))
	}
	return nil
}

func handleSSHCalendarsList(s ssh.Session) error {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return errors.New("calendar notifications are disabled")
	}
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle calendars list from ssh")

	if err := c.requireNotRevoked(s); err != nil {
		return err
	}
	usedRestriction, err := c.requireCapability(s, api.CapabilityNotifyAnyTokenRead)
	if err != nil {
		return err
	}

	umt := c.mt.ToUniversalMytoken()
	transactFn := func(tx *sqlx.Tx) (*model.Response, error) {
		infos, err := calendarrepo.List(c.rlog, tx, c.mt.ID)
		if err != nil {
			return nil, err
		}
		res := &model.Response{
			Status: http.StatusOK,
			Response: &calpkg.CalendarListResponse{
				CalendarListResponse: api.CalendarListResponse{Calendars: infos},
			},
		}
		return doAfterRequest(
			c.rlog, tx, res, c.mt, *c.clientMetaData, api.EventCalendarListed, "", usedRestriction, umt,
		)
	}
	res, err := transactWithResult(c.rlog, transactFn)
	if err != nil {
		return err
	}
	if res != nil {
		return writeJSON(s, res.Response)
	}
	return writeString(s, "[]")
}

func handleSSHCalendarCreate(reqData []byte, s ssh.Session) error {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return errors.New("calendar notifications are disabled")
	}
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle calendar-create from ssh")

	var req struct {
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}

	if err := c.requireNotRevoked(s); err != nil {
		return err
	}
	usedRestriction, err := c.requireCapability(s, api.CapabilityNotifyAnyToken)
	if err != nil {
		return err
	}

	id := utils.RandASCIIString(32)
	icsPath := calpkg.GetICSPath(id)
	calendarInfo := api.NotificationCalendar{
		ID:          id,
		ICSPath:     icsPath,
		Description: req.Description,
	}
	dbInfo := calendarrepo.CalendarInfo{
		ID:          id,
		Description: db.NewNullString(req.Description),
	}

	umt := c.mt.ToUniversalMytoken()
	transactFn := func(tx *sqlx.Tx) (*model.Response, error) {
		if err := calendarrepo.Insert(c.rlog, tx, c.mt.ID, dbInfo); err != nil {
			return nil, err
		}
		if len(req.Tags) > 0 {
			if err := calendarrepo.LinkTags(c.rlog, tx, id, req.Tags); err != nil {
				return nil, err
			}
		}
		res := &model.Response{
			Status: http.StatusCreated,
			Response: &calpkg.CreateCalendarResponse{
				NotificationCalendar: calendarInfo,
			},
		}
		return doAfterRequest(
			c.rlog, tx, res, c.mt, *c.clientMetaData, api.EventCalendarCreated, "", usedRestriction, umt,
		)
	}
	res, err := transactWithResult(c.rlog, transactFn)
	if err != nil {
		return err
	}
	if res != nil {
		return writeJSON(s, res.Response)
	}
	return writeString(s, "OK")
}

func handleSSHCalendarGet(reqData []byte, s ssh.Session) error {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return errors.New("calendar notifications are disabled")
	}
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle calendar-get from ssh")

	var req struct {
		CalendarID string `json:"calendar_id"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.CalendarID == "" {
		return errors.New("calendar_id is required")
	}

	if err := c.requireNotRevoked(s); err != nil {
		return err
	}

	if err := requireCalendarAccess(c.rlog, s, req.CalendarID, c.mt.ID); err != nil {
		return err
	}

	info, err := calendarrepo.GetByID(c.rlog, nil, req.CalendarID)
	if err != nil {
		return writeErrRes(s, model.ErrorToInternalServerErrorResponse(err))
	}
	resInfo, err := info.ToCalendarInfoResponse(c.rlog, nil)
	if err != nil {
		return writeErrRes(s, model.ErrorToInternalServerErrorResponse(err))
	}
	return writeJSON(s, resInfo)
}

func handleSSHCalendarDelete(reqData []byte, s ssh.Session) error {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return errors.New("calendar notifications are disabled")
	}
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle calendar-delete from ssh")

	var req struct {
		CalendarID string `json:"calendar_id"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.CalendarID == "" {
		return errors.New("calendar_id is required")
	}

	if err := c.requireNotRevoked(s); err != nil {
		return err
	}
	usedRestriction, err := c.requireCapability(s, api.CapabilityNotifyAnyToken)
	if err != nil {
		return err
	}

	umt := c.mt.ToUniversalMytoken()
	transactFn := func(tx *sqlx.Tx) (*model.Response, error) {
		if err := calendarrepo.Delete(c.rlog, tx, c.mt.ID, req.CalendarID); err != nil {
			return nil, err
		}
		return doAfterRequest(
			c.rlog, tx, nil, c.mt, *c.clientMetaData, api.EventCalendarDeleted, "", usedRestriction, umt,
		)
	}
	res, err := transactWithResult(c.rlog, transactFn)
	if err != nil {
		return err
	}
	if res != nil {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHCalendarAddMytoken(reqData []byte, s ssh.Session) error {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return errors.New("calendar notifications are disabled")
	}
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle calendar-add-mytoken from ssh")

	var req struct {
		CalendarID string `json:"calendar_id"`
		MomID      string `json:"mom_id"`
		Comment    string `json:"comment"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.CalendarID == "" {
		return errors.New("calendar_id is required")
	}

	if err := c.requireNotRevoked(s); err != nil {
		return err
	}

	if err := requireCalendarAccess(c.rlog, s, req.CalendarID, c.mt.ID); err != nil {
		return err
	}

	result, err := c.resolveMomMode(s, req.MomID, api.CapabilityTokeninfoNotify, api.CapabilityNotifyAnyToken)
	if err != nil {
		return err
	}

	umt := c.mt.ToUniversalMytoken()
	transactFn := func(tx *sqlx.Tx) (*model.Response, error) {
		info, err := calendarrepo.GetByID(c.rlog, tx, req.CalendarID)
		if err != nil {
			return nil, err
		}
		if err := calendarrepo.AddMytokenToCalendar(c.rlog, tx, result.id, info.ID); err != nil {
			return nil, err
		}
		event := api.EventNotificationSubscribed
		if result.momMode {
			event = api.EventNotificationSubscribedOther
		}
		resInfo, err := info.ToCalendarInfoResponse(c.rlog, tx)
		if err != nil {
			return nil, err
		}
		res := &model.Response{
			Status:   http.StatusOK,
			Response: resInfo,
		}
		return doAfterRequest(c.rlog, tx, res, c.mt, *c.clientMetaData, event, "", result.usedRestriction, umt)
	}
	res, err := transactWithResult(c.rlog, transactFn)
	if err != nil {
		return err
	}
	if res != nil {
		return writeJSON(s, res.Response)
	}
	return writeString(s, "OK")
}

func handleSSHCalendarAddTag(reqData []byte, s ssh.Session) error {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return errors.New("calendar notifications are disabled")
	}
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle calendar-add-tag from ssh")

	var req struct {
		CalendarID string `json:"calendar_id"`
		Tag        string `json:"tag"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.CalendarID == "" {
		return errors.New("calendar_id is required")
	}
	if req.Tag == "" {
		return errors.New("tag is required")
	}

	if err := c.requireNotRevoked(s); err != nil {
		return err
	}
	usedRestriction, err := c.requireCapability(s, api.CapabilityNotifyAnyToken)
	if err != nil {
		return err
	}

	if err := requireCalendarAccess(c.rlog, s, req.CalendarID, c.mt.ID); err != nil {
		return err
	}

	umt := c.mt.ToUniversalMytoken()
	transactFn := func(tx *sqlx.Tx) (*model.Response, error) {
		if err := calendarrepo.AddTag(c.rlog, tx, req.CalendarID, req.Tag); err != nil {
			return nil, err
		}
		return doAfterRequest(
			c.rlog, tx, nil, c.mt, *c.clientMetaData, api.EventCalendarListed, "", usedRestriction, umt,
		)
	}
	res, err := transactWithResult(c.rlog, transactFn)
	if err != nil {
		return err
	}
	if res != nil {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHCalendarRemoveTag(reqData []byte, s ssh.Session) error {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return errors.New("calendar notifications are disabled")
	}
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle calendar-remove-tag from ssh")

	var req struct {
		CalendarID string `json:"calendar_id"`
		Tag        string `json:"tag"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.CalendarID == "" {
		return errors.New("calendar_id is required")
	}
	if req.Tag == "" {
		return errors.New("tag is required")
	}

	if err := c.requireNotRevoked(s); err != nil {
		return err
	}
	usedRestriction, err := c.requireCapability(s, api.CapabilityNotifyAnyToken)
	if err != nil {
		return err
	}

	if err := requireCalendarAccess(c.rlog, s, req.CalendarID, c.mt.ID); err != nil {
		return err
	}

	umt := c.mt.ToUniversalMytoken()
	transactFn := func(tx *sqlx.Tx) (*model.Response, error) {
		if err := calendarrepo.RemoveTag(c.rlog, tx, req.CalendarID, req.Tag); err != nil {
			return nil, err
		}
		return doAfterRequest(
			c.rlog, tx, nil, c.mt, *c.clientMetaData, api.EventCalendarListed, "", usedRestriction, umt,
		)
	}
	res, err := transactWithResult(c.rlog, transactFn)
	if err != nil {
		return err
	}
	if res != nil {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHCalendarUpdate(reqData []byte, s ssh.Session) error {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return errors.New("calendar notifications are disabled")
	}
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle calendar-update from ssh")

	var req struct {
		CalendarID  string   `json:"calendar_id"`
		Description *string  `json:"description"`
		Tags        []string `json:"tags"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.CalendarID == "" {
		return errors.New("calendar_id is required")
	}
	if req.Description == nil && req.Tags == nil {
		return errors.New("no update parameters provided")
	}

	if err := c.requireNotRevoked(s); err != nil {
		return err
	}
	usedRestriction, err := c.requireCapability(s, api.CapabilityNotifyAnyToken)
	if err != nil {
		return err
	}

	if err := requireCalendarAccess(c.rlog, s, req.CalendarID, c.mt.ID); err != nil {
		return err
	}

	umt := c.mt.ToUniversalMytoken()
	transactFn := func(tx *sqlx.Tx) (*model.Response, error) {
		if req.Description != nil {
			if err := calendarrepo.UpdateDescription(
				c.rlog, tx, c.mt.ID, req.CalendarID, *req.Description,
			); err != nil {
				return nil, err
			}
		}
		if req.Tags != nil {
			if err := calendarrepo.LinkTags(c.rlog, tx, req.CalendarID, req.Tags); err != nil {
				return nil, err
			}
		}
		eventComment := "updated calendar"
		if req.Description == nil && req.Tags != nil {
			eventComment = "updated tags"
		} else if req.Description != nil && req.Tags == nil {
			eventComment = "updated description"
		}
		return doAfterRequest(
			c.rlog, tx, nil, c.mt, *c.clientMetaData, api.EventCalendarUpdated, eventComment, usedRestriction, umt,
		)
	}
	res, err := transactWithResult(c.rlog, transactFn)
	if err != nil {
		return err
	}
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHCalendarRemoveMytoken(reqData []byte, s ssh.Session) error {
	if !config.Get().Features.Notifications.ICS.Enabled {
		return errors.New("calendar notifications are disabled")
	}
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle calendar-remove-mytoken from ssh")

	var req struct {
		CalendarID string `json:"calendar_id"`
		MomID      string `json:"mom_id"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.CalendarID == "" {
		return errors.New("calendar_id is required")
	}

	if err := c.requireNotRevoked(s); err != nil {
		return err
	}

	if err := requireCalendarAccess(c.rlog, s, req.CalendarID, c.mt.ID); err != nil {
		return err
	}

	result, err := c.resolveMomMode(s, req.MomID, api.CapabilityTokeninfoNotify, api.CapabilityNotifyAnyToken)
	if err != nil {
		return err
	}

	umt := c.mt.ToUniversalMytoken()
	transactFn := func(tx *sqlx.Tx) (*model.Response, error) {
		if err := calendarrepo.RemoveMytokenFromCalendar(c.rlog, tx, result.id, req.CalendarID); err != nil {
			return nil, err
		}
		event := api.EventNotificationUnsubscribed
		if result.momMode {
			event = api.EventNotificationUnsubscribedOther
		}
		return doAfterRequest(c.rlog, tx, nil, c.mt, *c.clientMetaData, event, "", result.usedRestriction, umt)
	}
	res, err := transactWithResult(c.rlog, transactFn)
	if err != nil {
		return err
	}
	if res != nil {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}
