package notification

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
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
				Response: info,
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
			mtID := mtid.MOMID{MTID: mt.ID}
			var usedRestriction *restrictions.Restriction
			welcomeData := map[string]any{
				"management-url":       routes.NotificationManagementURL(managementCode),
				"token-name":           mt.Name, // in case of momid req we will replace it later
				"issuer-url":           config.Get().IssuerURL,
				"notification_classes": req.NotificationClasses,
			}
			requiredCapability := api.CapabilityTokeninfoNotify
			if req.MomID.HashValid() {
				mtID = req.MomID
				if res = auth.RequireMytokensForSameUser(rlog, tx, mtID.MTID, mt.ID); res != nil {
					return errors.New("rollback")
				}
				name, err := mytokenrepohelper.GetMTName(rlog, tx, mtID.MTID)
				if err != nil {
					return err
				}
				welcomeData["token-name"] = name.String
				requiredCapability = api.CapabilityNotifyAnyToken
			}
			usedRestriction, res = auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, ctxutils.ClientMetaData(ctx), requiredCapability,
			)
			if res != nil {
				return errors.New("rollback")
			}
			if !req.UserWide {
				welcomeData["mtid"] = mtID.Hash()
			}
			if err := notificationsrepo.NewNotification(rlog, tx, req, mtID, managementCode, ""); err != nil {
				return err
			}
			if err := usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
				return err
			}
			emailInfo, err := userrepo.GetMail(rlog, tx, mt.ID)
			if err != nil {
				return err
			}
			if !emailInfo.Mail.Valid {
				res = &model.Response{
					Status:   fiber.StatusUnprocessableEntity,
					Response: api.ErrorMailRequired,
				}
				return errors.New("rollback")
			}
			if !emailInfo.MailVerified {
				res = &model.Response{
					Status:   fiber.StatusUnprocessableEntity,
					Response: api.ErrorMailNotVerified,
				}
				return errors.New("rollback")
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
			return notificationsrepo.UpdateNotificationClasses(rlog, tx, info.NotificationID, req.Classes)
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
			return notificationsrepo.AddTokenToNotification(rlog, tx, info.NotificationID, mtID, req.IncludeChildren)
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
			return notificationsrepo.RemoveTokenFromNotification(rlog, tx, info.NotificationID, mtID)
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
