package mytokenutils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	pkg2 "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	pkg3 "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/rotation"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
)

// DoAfterRequestThingsOther does multiple things that are done typically after a request for "other" requests (
// no AT), in particular this means: token rotation, marks restriction usage, event logging
func DoAfterRequestThingsOther(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, resIn *model.Response, oldMT *mytoken.Mytoken,
	clientMetaData api.ClientMetaData,
	event api.Event, eventComment string,
	usedRestriction *restrictions.Restriction,
	oldJWT string, responseType model.ResponseType,
) (res *model.Response, rollback bool) {
	res = resIn
	rollback = true

	if usedRestriction != nil {
		if err := usedRestriction.UsedOther(rlog, tx, oldMT.ID); err != nil {
			res = model.ErrorToInternalServerErrorResponse(err)
		}
	}

	if event != api.EventUnknown {
		if err := eventService.LogEvent(
			rlog, tx, pkg3.MTEvent{
				Event:          event,
				Comment:        eventComment,
				MTID:           oldMT.ID,
				ClientMetaData: clientMetaData,
			},
		); err != nil {
			res = model.ErrorToInternalServerErrorResponse(err)
			return
		}
	}

	if oldJWT != "" {
		tokenUpdate, err := rotation.RotateMytokenAfterOtherForResponse(
			rlog, tx, oldJWT, oldMT, clientMetaData, responseType,
		)
		if err != nil {
			res = model.ErrorToInternalServerErrorResponse(err)
			return
		}
		if tokenUpdate != nil {
			if res == nil {
				res = &model.Response{Status: fiber.StatusOK}
			}
			if res.Status == fiber.StatusNoContent {
				res.Status = fiber.StatusOK
			}
			res.Cookies = []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)}
			tokenUpdatable, ok := res.Response.(pkg2.TokenUpdatableResponse)
			if ok {
				tokenUpdatable.SetTokenUpdate(tokenUpdate)
			} else {
				res.Response = pkg2.OnlyTokenUpdateRes{TokenUpdate: tokenUpdate}
			}
		}
	}
	rollback = false
	return
}
