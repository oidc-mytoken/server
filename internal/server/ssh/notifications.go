package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/endpoints/notification/pkg"
	"github.com/oidc-mytoken/server/internal/service/notification"
)

func handleSSHNotificationsList(s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle notifications list from ssh")
	res := notification.Service.List(c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData)
	if res == nil {
		return writeError(s, errors.New("internal server error"))
	}
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeJSON(s, res.Response)
}

func handleSSHNotificationCreate(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle notification-create from ssh")

	var req pkg.SubscribeNotificationRequest
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.NotificationType == "" {
		return errors.New("notification_type is required")
	}
	res := notification.Service.Create(c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData, req)
	if res == nil {
		return writeError(s, errors.New("internal server error"))
	}
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeJSON(s, res.Response)
}

func handleSSHNotificationAddToken(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle notification-add-token from ssh")

	var req struct {
		ManagementCode  string `json:"management_code"`
		MomID           string `json:"mom_id"`
		IncludeChildren bool   `json:"include_children"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.ManagementCode == "" {
		return errors.New("management_code is required")
	}
	res := notification.Service.AddToken(
		c.rlog, c.mt, c.clientMetaData, req.ManagementCode, req.MomID, req.IncludeChildren,
	)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHNotificationRemoveToken(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle notification-remove-token from ssh")

	var req struct {
		ManagementCode string `json:"management_code"`
		MomID          string `json:"mom_id"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.ManagementCode == "" {
		return errors.New("management_code is required")
	}
	res := notification.Service.RemoveToken(
		c.rlog, c.mt, c.clientMetaData, req.ManagementCode, req.MomID,
	)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

type notificationUpdateSSHRequest struct {
	ManagementCode string                   `json:"management_code"`
	Classes        *api.NotificationClasses `json:"classes,omitempty"`
	Tags           *[]api.Tag               `json:"tags,omitempty"`
}

func handleSSHNotificationUpdate(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle notification-update from ssh")

	var req notificationUpdateSSHRequest
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.ManagementCode == "" {
		return errors.New("management_code is required")
	}
	if req.Classes == nil && req.Tags == nil {
		return errors.New("no update parameters provided")
	}
	res := notification.Service.UpdateClasses(
		c.rlog, c.mt, c.clientMetaData, req.ManagementCode, req.Classes, req.Tags,
	)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHNotificationDelete(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle notification-delete from ssh")

	var req struct {
		ManagementCode string `json:"management_code"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.ManagementCode == "" {
		return errors.New("management_code is required")
	}
	res := notification.Service.DeleteByCode(c.rlog, c.mt, c.clientMetaData, req.ManagementCode)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}
