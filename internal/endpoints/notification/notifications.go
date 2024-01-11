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
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	pkg3 "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/rotation"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	notifier "github.com/oidc-mytoken/server/internal/notifier/client"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleGetByManagementCode returns the api.NotificationInfo for the notification linked to a management code
func HandleGetByManagementCode(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle notification get request for management code")
	managementCode := ctx.Params("code")
	if managementCode == "" {
		return model.BadRequestErrorResponse("missing management_code").Send(ctx)
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
				res = model.NotFoundErrorResponse("management_code not valid")
				return errors.New("dummy")
			}
			res = &model.Response{
				Status:   fiber.StatusOK,
				Response: info,
			}
			return nil
		},
	)
	return res.Send(ctx)
}

// HandleDeleteByManagementCode returns the api.NotificationInfo for the notification linked to a management code
func HandleDeleteByManagementCode(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle notification get request for management code")
	managementCode := ctx.Params("code")
	if managementCode == "" {
		return model.BadRequestErrorResponse("missing management_code").Send(ctx)
	}

	err := notificationsrepo.Delete(rlog, nil, managementCode)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return model.Response{Status: fiber.StatusNoContent}.Send(ctx)
}

// HandleGet handles get requests and returns a list of all notifications for a user
func HandleGet(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle notification get request")
	var umt universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &umt, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt,
		ctxutils.ClientMetaData(ctx), api.CapabilityNotifyAnyTokenRead,
	)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	var res *model.Response
	_ = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			infos, err := notificationsrepo.GetNotificationsForUser(rlog, tx, mt.ID)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			resData := pkg.NotificationsListResponse{
				NotificationsListResponse: api.NotificationsListResponse{
					Notifications: infos,
				},
			}
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
				rlog, tx, pkg3.MTEvent{
					Event:          api.EventNotificationListed,
					MTID:           mt.ID,
					ClientMetaData: *ctxutils.ClientMetaData(ctx),
				},
			)
		},
	)
	return res.Send(ctx)
}

// HandlePost is the main entry function for handling notification creation requests
func HandlePost(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle notification create request")
	var req pkg.SubscribeNotificationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &req.Mytoken, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	managementCode := utils.RandASCIIString(64)
	switch req.NotificationType {
	case api.NotificationTypeICSInvite:
		return calendar.HandleCalendarEntryViaMail(ctx, rlog, mt, req)
	case api.NotificationTypeMail:
		return handleNewMailNotification(ctx, rlog, mt, req, managementCode)
	case api.NotificationTypeWebsocket:
		return model.ResponseNYI.Send(ctx)
	default:
		return model.BadRequestErrorResponse("unknown notification_type").Send(ctx)
	}
}

func handleNewMailNotification(
	ctx *fiber.Ctx, rlog logrus.Ext1FieldLogger, mt *mytoken.Mytoken,
	req pkg.SubscribeNotificationRequest, managementCode string,
) error {
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			mtID := mtid.MOMID{MTID: mt.ID}
			var usedRestriction *restrictions.Restriction
			welcomeData := map[string]any{
				"management-url":       managementCode, //TODO
				"token-name":           mt.Name,        // in case of momid req we will replace it later
				"issuer-url":           config.Get().IssuerURL,
				"notification_classes": req.NotificationClasses,
			}
			if req.MomID.HashValid() {
				mtID = req.MomID
				name, err := mytokenrepohelper.GetMTName(rlog, tx, mtID.MTID)
				if err != nil {
					return err
				}
				welcomeData["token-name"] = name.String
				usedRestriction, res = auth.RequireCapabilityAndRestrictionOther(
					rlog, nil, mt, ctxutils.ClientMetaData(ctx), api.CapabilityNotifyAnyToken,
				)
				if res != nil {
					return errors.New("dummy")
				}
			} else { // mytoken notification for itself
				usedRestriction, res = auth.RequireCapabilityAndRestrictionOther(
					rlog, nil, mt, ctxutils.ClientMetaData(ctx), api.CapabilityTokeninfoNotify,
				)
				if res != nil {
					return errors.New("dummy")
				}
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
			notifier.SendTemplateEmail(
				emailInfo.Mail, "New Mytoken Notification Subscription",
				emailInfo.PreferHTMLMail, "notification-welcome", welcomeData,
			)
			res = &model.Response{
				Status:   fiber.StatusCreated,
				Response: managementCode, //TODO
			}
			return nil
		},
	); err != nil {
		if res != nil {
			return res.Send(ctx)
		}
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return res.Send(ctx)
}

// HandleNotificationUpdateClasses handles requests to update the NotificationClasses for a notification
func HandleNotificationUpdateClasses(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle notification update classes request")
	managementCode := ctx.Params("code")
	if managementCode == "" {
		return model.BadRequestErrorResponse("missing management_code").Send(ctx)
	}
	var req api.NotificationUpdateNotificationClassesRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			info, err := notificationsrepo.GetNotificationForManagementCode(rlog, tx, managementCode)
			if err != nil {
				return err
			}
			if info == nil {
				res = model.NotFoundErrorResponse("management_code not valid")
				return nil
			}
			return notificationsrepo.UpdateNotificationClasses(rlog, tx, info.NotificationID, req.Classes)
		},
	)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	if res == nil {
		res = &model.Response{Status: fiber.StatusNoContent}
	}
	return res.Send(ctx)
}

// HandleNotificationAddToken handles requests to add a mytoken to a notification
func HandleNotificationAddToken(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle notification add token request")
	managementCode := ctx.Params("code")
	if managementCode == "" {
		return model.BadRequestErrorResponse("missing management_code").Send(ctx)
	}
	var req pkg.NotificationAddTokenRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			mtID := req.MomID
			if !mtID.HashValid() {
				mt, errRes := auth.RequireValidMytoken(rlog, tx, &req.Mytoken, nil)
				if errRes != nil {
					res = errRes
					return nil
				}
				mtID = mtid.MOMID{MTID: mt.ID}
			}
			info, err := notificationsrepo.GetNotificationForManagementCode(rlog, tx, managementCode)
			if err != nil {
				return err
			}
			if info == nil {
				res = model.NotFoundErrorResponse("management_code not valid")
				return nil
			}
			return notificationsrepo.AddTokenToNotification(rlog, tx, info.NotificationID, mtID, req.IncludeChildren)
		},
	)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	if res == nil {
		res = &model.Response{Status: fiber.StatusNoContent}
	}
	return res.Send(ctx)
}

// HandleNotificationRemoveToken handles requests to remove a mytoken from a notification
func HandleNotificationRemoveToken(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle notification remove token request")
	managementCode := ctx.Params("code")
	if managementCode == "" {
		return model.BadRequestErrorResponse("missing management_code").Send(ctx)
	}
	var req pkg.NotificationRemoveTokenRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			mtID := req.MomID
			if !mtID.HashValid() {
				mt, errRes := auth.RequireValidMytoken(rlog, tx, &req.Mytoken, nil)
				if errRes != nil {
					res = errRes
					return nil
				}
				mtID = mtid.MOMID{MTID: mt.ID}
			}
			info, err := notificationsrepo.GetNotificationForManagementCode(rlog, tx, managementCode)
			if err != nil {
				return err
			}
			if info == nil {
				res = model.NotFoundErrorResponse("management_code not valid")
				return nil
			}
			return notificationsrepo.RemoveTokenFromNotification(rlog, tx, info.NotificationID, mtID)
		},
	)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	if res == nil {
		res = &model.Response{Status: fiber.StatusNoContent}
	}
	return res.Send(ctx)
}
