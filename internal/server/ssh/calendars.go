package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/service/calendar"
)

func handleSSHCalendarsList(s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle calendars list from ssh")
	res := calendar.Service.List(c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	if res != nil && res.Response != nil {
		return writeJSON(s, res.Response)
	}
	return writeString(s, "[]")
}

func handleSSHCalendarCreate(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle calendar-create from ssh")

	var req struct {
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}

	res := calendar.Service.Create(c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData, req.Description, req.Tags)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	if res != nil && res.Response != nil {
		return writeJSON(s, res.Response)
	}
	return writeString(s, "OK")
}

func handleSSHCalendarGet(reqData []byte, s ssh.Session) error {
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

	res := calendar.Service.Get(c.rlog, c.mt, c.clientMetaData, req.CalendarID)
	if res == nil {
		return errors.New("unexpected nil response")
	}
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeJSON(s, res.Response)
}

func handleSSHCalendarDelete(reqData []byte, s ssh.Session) error {
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

	res := calendar.Service.Delete(c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData, req.CalendarID)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHCalendarAddMytoken(reqData []byte, s ssh.Session) error {
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

	momID := c.mt.ID.MomID()
	if req.MomID != "" {
		momID = mtid.MOMID{MTID: mtid.FromHash(req.MomID)}
	}

	res := calendar.Service.AddMytoken(
		c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData, req.CalendarID, req.Comment, momID,
	)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	if res != nil && res.Response != nil {
		return writeJSON(s, res.Response)
	}
	return writeString(s, "OK")
}

func handleSSHCalendarRemoveMytoken(reqData []byte, s ssh.Session) error {
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

	momID := c.mt.ID.MomID()
	if req.MomID != "" {
		momID = mtid.MOMID{MTID: mtid.FromHash(req.MomID)}
	}

	res := calendar.Service.RemoveMytoken(
		c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData, req.CalendarID, momID,
	)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHCalendarAddTag(reqData []byte, s ssh.Session) error {
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

	res := calendar.Service.AddTag(c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData, req.CalendarID, req.Tag)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHCalendarRemoveTag(reqData []byte, s ssh.Session) error {
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

	res := calendar.Service.RemoveTag(
		c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData, req.CalendarID, req.Tag,
	)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHCalendarUpdate(reqData []byte, s ssh.Session) error {
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

	res := calendar.Service.Update(
		c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData, req.CalendarID, req.Description, req.Tags,
	)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}
