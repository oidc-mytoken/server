package notification

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/notification/calendar"
	"github.com/oidc-mytoken/server/internal/endpoints/notification/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

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
		return model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError("unknown notification_type"),
		}.Send(ctx)
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
			if req.MomID.Valid() {
				mtID = req.MomID
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
			if err := notificationsrepo.NewNotification(rlog, nil, req, mtID, managementCode, ""); err != nil {
				return err
			}
			if err := usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
				return err
			}
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
