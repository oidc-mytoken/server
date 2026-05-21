package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/notification/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/internal/utils/mytokenutils"
)

func handleSSHNotificationsList(s ssh.Session) error {
	if !config.Get().Features.Notifications.AnyEnabled {
		return errors.New("notifications are disabled")
	}
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle notifications list from ssh")

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt, clientMetaData, api.CapabilityNotifyAnyTokenRead,
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}

	umt := mt.ToUniversalMytoken()
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			infos, err := notificationsrepo.GetNotificationsForUser(rlog, tx, mt.ID)
			if err != nil {
				return err
			}
			res = &model.Response{
				Status: 200,
				Response: &pkg.NotificationsListResponse{
					NotificationsListResponse: api.NotificationsListResponse{
						Notifications: infos,
					},
				},
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, res, mt, *clientMetaData,
				api.EventNotificationListed, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	)
	if err != nil && res == nil {
		return err
	}
	if res != nil {
		return writeJSON(s, res.Response)
	}
	return writeString(s, "[]")
}

func handleSSHNotificationCreate(reqData []byte, s ssh.Session) error {
	if !config.Get().Features.Notifications.AnyEnabled {
		return errors.New("notifications are disabled")
	}
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle notification-create from ssh")

	var req struct {
		pkg.SubscribeNotificationRequest
		MomID string `json:"mom_id"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.NotificationType == "" {
		return errors.New("notification_type is required")
	}

	managementCode := utils.RandASCIIString(64)
	mtID := mt.ID.MomID()
	requiredCapability := api.CapabilityTokeninfoNotify
	if req.MomID != "" {
		mtID = mtid.MOMID{MTID: mtid.FromHash(req.MomID)}
		requiredCapability = api.CapabilityNotifyAnyToken
	}

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt, clientMetaData, requiredCapability,
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}

	umt := mt.ToUniversalMytoken()
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err := notificationsrepo.NewNotification(
				rlog, tx, req.SubscribeNotificationRequest, mtID, managementCode, "",
			); err != nil {
				return err
			}
			res = &model.Response{
				Status: 201,
				Response: &pkg.NotificationsCreateResponse{
					NotificationsCreateResponse: api.NotificationsCreateResponse{
						ManagementCode: managementCode,
					},
				},
			}
			e := api.EventNotificationCreated
			if req.MomID != "" {
				e = api.EventNotificationCreatedOther
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, res, mt, *clientMetaData, e, "",
				usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	)
	if err != nil && res == nil {
		return err
	}
	if res != nil {
		return writeJSON(s, res.Response)
	}
	return writeString(s, "OK")
}

func handleSSHNotificationAddToken(reqData []byte, s ssh.Session) error {
	if !config.Get().Features.Notifications.AnyEnabled {
		return errors.New("notifications are disabled")
	}
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle notification-add-token from ssh")

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

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}

	mtID := mt.ID
	if req.MomID != "" {
		mtID = mtid.FromHash(req.MomID)
	}
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			info, err := notificationsrepo.GetNotificationForManagementCode(rlog, tx, req.ManagementCode)
			if err != nil {
				return err
			}
			if info == nil {
				res = model.NotFoundErrorResponse("management_code not valid")
				return errors.New("rollback")
			}
			momID := mtid.MOMID{MTID: mtID}
			if err = notificationsrepo.AddTokenToNotification(
				rlog, tx, info.NotificationID, momID, req.IncludeChildren,
			); err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil && res == nil {
		res = model.ErrorToInternalServerErrorResponse(err)
	}
	if res != nil {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHNotificationRemoveToken(reqData []byte, s ssh.Session) error {
	if !config.Get().Features.Notifications.AnyEnabled {
		return errors.New("notifications are disabled")
	}
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle notification-remove-token from ssh")

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

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}

	mtID := mt.ID
	if req.MomID != "" {
		mtID = mtid.FromHash(req.MomID)
	}
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			info, err := notificationsrepo.GetNotificationForManagementCode(rlog, tx, req.ManagementCode)
			if err != nil {
				return err
			}
			if info == nil {
				res = model.NotFoundErrorResponse("management_code not valid")
				return errors.New("rollback")
			}
			momID := mtid.MOMID{MTID: mtID}
			if err = notificationsrepo.RemoveTokenFromNotification(rlog, tx, info.NotificationID, momID); err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil && res == nil {
		res = model.ErrorToInternalServerErrorResponse(err)
	}
	if res != nil {
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
	if !config.Get().Features.Notifications.AnyEnabled {
		return errors.New("notifications are disabled")
	}
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle notification-update from ssh")

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

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}

	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			info, err := notificationsrepo.GetNotificationForManagementCode(rlog, tx, req.ManagementCode)
			if err != nil {
				return err
			}
			if info == nil {
				res = model.NotFoundErrorResponse("management_code not valid")
				return errors.New("rollback")
			}
			if req.Classes != nil {
				if err = notificationsrepo.UpdateNotificationClasses(
					rlog, tx, info.NotificationID, *req.Classes,
				); err != nil {
					return err
				}
			}
			if req.Tags != nil {
				if err = notificationsrepo.LinkTags(rlog, tx, info.NotificationID, *req.Tags); err != nil {
					return err
				}
			}
			return nil
		},
	)
	if err != nil && res == nil {
		res = model.ErrorToInternalServerErrorResponse(err)
	}
	if res != nil {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHNotificationDelete(reqData []byte, s ssh.Session) error {
	if !config.Get().Features.Notifications.AnyEnabled {
		return errors.New("notifications are disabled")
	}
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle notification-delete from ssh")

	var req struct {
		ManagementCode string `json:"management_code"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.ManagementCode == "" {
		return errors.New("management_code is required")
	}

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}

	err := notificationsrepo.Delete(rlog, nil, req.ManagementCode)
	if err != nil {
		return err
	}
	return writeString(s, "OK")
}
