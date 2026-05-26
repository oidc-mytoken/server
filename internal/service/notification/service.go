// Package notification provides a service layer for notification operations.
package notification

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/notification/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/mytokenutils"
)

// Service is the notification service singleton.
var Service = &service{}

type service struct{}

// List returns all notifications for the user.
func (s *service) List(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData,
) *model.Response {
	if !config.Get().Features.Notifications.AnyEnabled {
		return model.BadRequestErrorResponse("notifications are disabled")
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
			infos, err := notificationsrepo.GetNotificationsForUser(rlog, tx, mt.ID)
			if err != nil {
				return err
			}
			res = &model.Response{
				Status: http.StatusOK,
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
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

// Create creates a new notification.
func (s *service) Create(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, req pkg.SubscribeNotificationRequest,
) *model.Response {
	if !config.Get().Features.Notifications.AnyEnabled {
		return model.BadRequestErrorResponse("notifications are disabled")
	}
	managementCode := utils.RandASCIIString(64)
	mtID := mt.ID.MomID()
	requiredCapability := api.CapabilityTokeninfoNotify
	if req.MomID.HashValid() {
		mtID = req.MomID
		requiredCapability = api.CapabilityNotifyAnyToken
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, clientMetaData, requiredCapability,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if err := notificationsrepo.NewNotification(rlog, tx, req, mtID, managementCode, ""); err != nil {
				return err
			}
			res = &model.Response{
				Status: http.StatusCreated,
				Response: &pkg.NotificationsCreateResponse{
					NotificationsCreateResponse: api.NotificationsCreateResponse{
						ManagementCode: managementCode,
					},
				},
			}
			e := api.EventNotificationCreated
			if req.MomID.HashValid() {
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
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

// AddToken adds a mytoken to a notification identified by management code.
func (s *service) AddToken(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData, managementCode, momID string, includeChildren bool,
) *model.Response {
	if !config.Get().Features.Notifications.AnyEnabled {
		return model.BadRequestErrorResponse("notifications are disabled")
	}
	if managementCode == "" {
		return model.BadRequestErrorResponse("management_code is required")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			mtID := mt.ID
			if momID != "" {
				mtID = mtid.FromHash(momID)
			}
			info, err := notificationsrepo.GetNotificationForManagementCode(rlog, tx, managementCode)
			if err != nil {
				return err
			}
			if info == nil {
				res = model.NotFoundErrorResponse("management_code not valid")
				return errors.New("rollback")
			}
			momIDObj := mtid.MOMID{MTID: mtID}
			if err = notificationsrepo.AddTokenToNotification(
				rlog, tx, info.NotificationID, momIDObj, includeChildren,
			); err != nil {
				return err
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

// RemoveToken removes a mytoken from a notification.
func (s *service) RemoveToken(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData, managementCode, momID string,
) *model.Response {
	if !config.Get().Features.Notifications.AnyEnabled {
		return model.BadRequestErrorResponse("notifications are disabled")
	}
	if managementCode == "" {
		return model.BadRequestErrorResponse("management_code is required")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			mtID := mt.ID
			if momID != "" {
				mtID = mtid.FromHash(momID)
			}
			info, err := notificationsrepo.GetNotificationForManagementCode(rlog, tx, managementCode)
			if err != nil {
				return err
			}
			if info == nil {
				res = model.NotFoundErrorResponse("management_code not valid")
				return errors.New("rollback")
			}
			momIDObj := mtid.MOMID{MTID: mtID}
			if err = notificationsrepo.RemoveTokenFromNotification(
				rlog, tx, info.NotificationID, momIDObj,
			); err != nil {
				return err
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

// UpdateClasses updates the classes and/or tags of a notification.
func (s *service) UpdateClasses(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData, managementCode string,
	classes *api.NotificationClasses, tags *[]api.Tag,
) *model.Response {
	if !config.Get().Features.Notifications.AnyEnabled {
		return model.BadRequestErrorResponse("notifications are disabled")
	}
	if managementCode == "" {
		return model.BadRequestErrorResponse("management_code is required")
	}
	if classes == nil && tags == nil {
		return model.BadRequestErrorResponse("no update parameters provided")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			info, err := notificationsrepo.GetNotificationForManagementCode(rlog, tx, managementCode)
			if err != nil {
				return err
			}
			if info == nil {
				res = model.NotFoundErrorResponse("management_code not valid")
				return errors.New("rollback")
			}
			if classes != nil {
				if err = notificationsrepo.UpdateNotificationClasses(
					rlog, tx, info.NotificationID, *classes,
				); err != nil {
					return err
				}
			}
			if tags != nil {
				if err = notificationsrepo.LinkTags(rlog, tx, info.NotificationID, *tags); err != nil {
					return err
				}
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

// DeleteByCode deletes a notification by its management code.
func (s *service) DeleteByCode(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData, managementCode string,
) *model.Response {
	if !config.Get().Features.Notifications.AnyEnabled {
		return model.BadRequestErrorResponse("notifications are disabled")
	}
	if managementCode == "" {
		return model.BadRequestErrorResponse("management_code is required")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			err := notificationsrepo.Delete(rlog, tx, managementCode)
			if err != nil {
				return err
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
