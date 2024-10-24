package notification

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/unixtime"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/userrepo"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/notification/calendar"
	"github.com/oidc-mytoken/server/internal/endpoints/notification/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	notifier "github.com/oidc-mytoken/server/internal/notifier/client"
	"github.com/oidc-mytoken/server/internal/server/routes"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/internal/utils/mytokenutils"
)

var managementCodeNotValidError = model.NotFoundErrorResponse("management_code not valid")
var missingManagementCodeError = model.BadRequestErrorResponse("missing management_code")

// HandleGetByManagementCode returns the api.NotificationInfo for the notification linked to a management code
func HandleGetByManagementCode(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle notification get request for management code")
	managementCode := ctx.Params("code")
	if managementCode == "" {
		return missingManagementCodeError
	}

	var res *model.Response
	_ = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			info, err := notificationsrepo.GetNotificationForManagementCode(rlog, tx, managementCode)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if info == nil {
				res = managementCodeNotValidError
				return errors.New("rollback")
			}
			res = &model.Response{
				Status:   fiber.StatusOK,
				Response: info.ManagementCodeNotificationInfoResponse,
			}
			return nil
		},
	)
	return res
}

// HandleDeleteByManagementCode returns the api.NotificationInfo for the notification linked to a management code
func HandleDeleteByManagementCode(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle notification get request for management code")
	managementCode := ctx.Params("code")
	if managementCode == "" {
		return missingManagementCodeError
	}

	err := notificationsrepo.Delete(rlog, nil, managementCode)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{Status: fiber.StatusNoContent}
}

// HandleGet handles get requests and returns a list of all notifications for a user
func HandleGet(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle notification get request")
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
			infos, err := notificationsrepo.GetNotificationsForUser(rlog, tx, mt.ID)
			if err != nil {
				return err
			}
			res = &model.Response{
				Status: fiber.StatusOK,
				Response: &pkg.NotificationsListResponse{
					NotificationsListResponse: api.NotificationsListResponse{
						Notifications: infos,
					},
				},
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, res, mt, *ctxutils.ClientMetaData(ctx),
				api.EventNotificationListed, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
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

// HandlePost is the main entry function for handling notification creation requests
func HandlePost(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle notification create request")
	var req pkg.SubscribeNotificationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &req.Mytoken, ctx)
	if errRes != nil {
		return errRes
	}
	managementCode := utils.RandASCIIString(64)
	switch req.NotificationType {
	case api.NotificationTypeICSInvite:
		return calendar.HandleCalendarEntryViaMail(ctx, rlog, mt, req)
	case api.NotificationTypeMail:
		return handleNewMailNotification(ctx, rlog, mt, req, managementCode)
	case api.NotificationTypeWebsocket:
		return &model.ResponseNYI
	default:
		return model.BadRequestErrorResponse("unknown notification_type")
	}
}

func handleNewMailNotification(
	ctx *fiber.Ctx, rlog logrus.Ext1FieldLogger, mt *mytoken.Mytoken,
	req pkg.SubscribeNotificationRequest, managementCode string,
) *model.Response {
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			mtID, requiredCapability, welcomeData, errRes, err := prepareNotificationWelcomeData(
				rlog, tx, mt, req, managementCode,
			)
			if err != nil {
				res = errRes
				return err
			}
			var usedRestriction *restrictions.Restriction
			usedRestriction, res = auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, ctxutils.ClientMetaData(ctx), requiredCapability,
			)
			if res != nil {
				return errors.New("rollback")
			}
			if err = notificationsrepo.NewNotification(rlog, tx, req, mtID, managementCode, ""); err != nil {
				return err
			}
			if req.NotificationClasses.Contains(api.NotificationClassExpiration) {
				var withClass []notificationsrepo.NotificationInfoBaseWithClass
				if err = tx.Select(
					&withClass, `CALL Notifications_GetForManagementCode(?)`, managementCode,
				); err != nil {
					return err
				}
				if err = notificationsrepo.AddScheduledExpirationNotifications(
					rlog, tx, withClass[0].NotificationInfoBase,
				); err != nil {
					return err
				}
			}
			emailInfo, errRes, err := userrepo.GetAndCheckMail(rlog, tx, mt.ID)
			if err != nil {
				res = errRes
				return err
			}

			notifier.SendTemplateEmail(
				emailInfo.Mail.String, "New Mytoken Notification Subscription",
				emailInfo.PreferHTMLMail, "notification-welcome", welcomeData,
			)

			res = &model.Response{
				Status: fiber.StatusCreated,
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
				rlog, tx, res, mt, *ctxutils.ClientMetaData(ctx), e, "",
				usedRestriction, req.Mytoken.JWT, req.Mytoken.OriginalTokenType,
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

func prepareNotificationWelcomeData(
	rlog logrus.Ext1FieldLogger, tx *sqlx.Tx, mt *mytoken.Mytoken,
	req pkg.SubscribeNotificationRequest, managementCode string,
) (mtid.MOMID, api.Capability, map[string]interface{}, *model.Response, error) {
	mtID := mtid.MOMID{MTID: mt.ID}
	welcomeData := map[string]interface{}{
		"management-url":       routes.NotificationManagementURL(managementCode),
		"token-name":           mt.Name,
		"issuer-url":           config.Get().IssuerURL,
		"notification_classes": req.NotificationClasses,
	}
	requiredCapability := api.CapabilityTokeninfoNotify

	if req.MomID.HashValid() {
		mtID = req.MomID
		if errRes := auth.RequireMytokensForSameUser(rlog, tx, mtID.MTID, mt.ID); errRes != nil {
			return mtID, requiredCapability, welcomeData, errRes, errors.New("rollback")
		}
		name, err := mytokenrepohelper.GetMTName(rlog, tx, mtID.MTID)
		if err != nil {
			return mtID, requiredCapability, welcomeData, nil, err
		}
		welcomeData["token-name"] = name.String
		requiredCapability = api.CapabilityNotifyAnyToken
	}
	if !req.UserWide {
		welcomeData["mtid"] = mtID.Hash()
	}

	return mtID, requiredCapability, welcomeData, nil, nil
}

// HandleNotificationUpdateClasses handles requests to update the NotificationClasses for a notification
func HandleNotificationUpdateClasses(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle notification update classes request")
	managementCode := ctx.Params("code")
	if managementCode == "" {
		return missingManagementCodeError
	}
	var req api.NotificationUpdateNotificationClassesRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			info, err := notificationsrepo.GetNotificationForManagementCode(rlog, tx, managementCode)
			if err != nil {
				return err
			}
			if info == nil {
				res = managementCodeNotValidError
				return errors.New("rollback")
			}
			if err = notificationsrepo.UpdateNotificationClasses(
				rlog, tx, info.NotificationID,
				req.Classes,
			); err != nil {
				return err
			}
			includedExpBefore := info.Classes.Contains(api.NotificationClassExpiration)
			includesExpNow := req.Classes.Contains(api.NotificationClassExpiration)
			if includesExpNow && !includedExpBefore {
				// exp class was added
				if err = notificationsrepo.AddScheduledExpirationNotifications(
					rlog, tx,
					notificationsrepo.NotificationInfoBase{
						NotificationInfoBase: info.NotificationInfoBase,
						WebSocketPath:        db.NewNullString(info.WebSocketPath),
						UID:                  info.UID,
					},
				); err != nil {
					return err
				}
			}
			if includedExpBefore && !includesExpNow {
				// exp class was removed
				if err = notificationsrepo.DeleteScheduledExpirationNotifications(
					rlog, tx, info.NotificationID,
				); err != nil {
					return err
				}
			}
			return nil
		},
	)
	if err != nil && res == nil {
		res = model.ErrorToInternalServerErrorResponse(err)
	}
	if res == nil {
		res = &model.Response{Status: fiber.StatusNoContent}
	}
	return res
}

// HandleNotificationAddToken handles requests to add a mytoken to a notification
func HandleNotificationAddToken(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle notification add token request")
	managementCode := ctx.Params("code")
	if managementCode == "" {
		return missingManagementCodeError
	}
	var req pkg.NotificationAddTokenRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			mtID := req.MomID
			var expiresAt unixtime.UnixTime
			var createdAt unixtime.UnixTime
			if mtID.HashValid() {
				mtInfo, err := tree.SingleTokenEntry(rlog, tx, mtID.MTID)
				if err != nil {
					return err
				}
				expiresAt = mtInfo.ExpiresAt
				createdAt = mtInfo.CreatedAt
			} else {
				mt, errRes := auth.RequireValidMytoken(rlog, tx, &req.Mytoken, nil)
				if errRes != nil {
					res = errRes
					return errors.New("rollback")
				}
				mtID = mtid.MOMID{MTID: mt.ID}
				expiresAt = mt.ExpiresAt
				createdAt = mt.IssuedAt
			}
			info, err := notificationsrepo.GetNotificationForManagementCode(rlog, tx, managementCode)
			if err != nil {
				return err
			}
			if info == nil {
				res = managementCodeNotValidError
				return errors.New("rollback")
			}
			if err = notificationsrepo.AddTokenToNotification(
				rlog, tx, info.NotificationID, mtID,
				req.IncludeChildren,
			); err != nil {
				return err
			}
			if info.Classes.Contains(api.NotificationClassExpiration) {
				if err = notificationsrepo.ScheduleExpirationNotifications(
					rlog, tx, info.NotificationID, mtID.MTID, expiresAt, createdAt,
				); err != nil {
					return err
				}
			}
			return nil
		},
	)
	if err != nil && res == nil {
		res = model.ErrorToInternalServerErrorResponse(err)
	}
	if res == nil {
		res = &model.Response{Status: fiber.StatusNoContent}
	}
	return res
}

// HandleNotificationRemoveToken handles requests to remove a mytoken from a notification
func HandleNotificationRemoveToken(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle notification remove token request")
	managementCode := ctx.Params("code")
	if managementCode == "" {
		return missingManagementCodeError
	}
	var req pkg.NotificationRemoveTokenRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			mtID := req.MomID
			if !mtID.HashValid() {
				mt, errRes := auth.RequireValidMytoken(rlog, tx, &req.Mytoken, nil)
				if errRes != nil {
					res = errRes
					return errors.New("rollback")
				}
				mtID = mtid.MOMID{MTID: mt.ID}
			}
			info, err := notificationsrepo.GetNotificationForManagementCode(rlog, tx, managementCode)
			if err != nil {
				return err
			}
			if info == nil {
				res = managementCodeNotValidError
				return errors.New("rollback")
			}
			if err = notificationsrepo.RemoveTokenFromNotification(rlog, tx, info.NotificationID, mtID); err != nil {
				return err
			}
			return notificationsrepo.DeleteScheduledExpirationNotificationsForMT(
				rlog, tx, info.NotificationID, mtID.MTID,
			)
		},
	)
	if err != nil && res == nil {
		res = model.ErrorToInternalServerErrorResponse(err)
	}
	if res == nil {
		res = &model.Response{Status: fiber.StatusNoContent}
	}
	return res
}
