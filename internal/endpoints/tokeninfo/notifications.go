package tokeninfo

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	helper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	pkg2 "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/rotation"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

func doTokenInfoNotifications(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken,
	clientMetadata *api.ClientMetaData,
	usedRestriction *restrictions.Restriction,
) (res pkg.TokeninfoNotificationsResponse, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if len(req.MOMIDs) > 1 {
				res.MomIDMapping = make(map[string]api.NotificationsCombinedResponse)
				for _, id := range req.MOMIDs {
					if id == api.MOMIDValueThis {
						id = mt.ID.String()
					}
					var data api.NotificationsCombinedResponse
					data.Notifications, data.Calendars, err = notificationsrepo.GetNotificationsAndCalendarsForMT(
						rlog, tx, id,
					)
					if err != nil {
						return err
					}
					res.MomIDMapping[id] = data
				}
			} else {
				var id any
				id = mt.ID
				if len(req.MOMIDs) > 0 {
					id = req.MOMIDs[0]
				}
				res.Notifications, res.Calendars, err = notificationsrepo.GetNotificationsAndCalendarsForMT(
					rlog, tx, id,
				)
				if err != nil {
					return err
				}
			}
			if usedRestriction == nil {
				return nil
			}
			if err = usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
				return err
			}
			res.TokenUpdate, err = rotation.RotateMytokenAfterOtherForResponse(
				rlog, tx, req.Mytoken.JWT, mt, *clientMetadata, req.Mytoken.OriginalTokenType,
			)
			if err != nil {
				return err
			}
			ev := api.EventTokenInfoNotifications
			if len(req.MOMIDs) > 0 {
				ev = api.EventTokenInfoNotificationsOtherToken
			}
			return eventService.LogEvent(
				rlog, tx, pkg2.MTEvent{
					Event:          ev,
					MTID:           mt.ID,
					ClientMetaData: *clientMetadata,
				},
			)
		},
	)
	return
}

func handleTokenInfoNotifications(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken,
	clientMetadata *api.ClientMetaData,
) *model.Response {
	usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, nil, mt, clientMetadata)
	if errRes != nil {
		return errRes
	}
	res, err := doTokenInfoNotifications(rlog, tx, req, mt, clientMetadata, usedRestriction)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	rsp := &model.Response{
		Status:   fiber.StatusOK,
		Response: res,
	}
	if res.TokenUpdate != nil {
		rsp.Cookies = []*fiber.Cookie{cookies.MytokenCookie(res.TokenUpdate.Mytoken)}
	}
	return rsp
}

// HandleTokenInfoNotifications handles a tokeninfo notifications request
func HandleTokenInfoNotifications(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken,
	clientMetadata *api.ClientMetaData,
) *model.Response {
	// If we call this function it means the token is valid.

	rlog.Debug("Handle tokeninfo notifications request")
	if len(req.MOMIDs) == 0 {
		if errRes := auth.RequireCapability(
			rlog, tx, api.CapabilityTokeninfoNotify, mt, clientMetadata,
		); errRes != nil {
			return errRes
		}
		return handleTokenInfoNotifications(rlog, tx, req, mt, clientMetadata)
	}
	for _, momid := range req.MOMIDs {
		if !mt.Capabilities.Has(api.CapabilityNotifyAnyTokenRead) {
			if momid == api.MOMIDValueThis {
				continue
			}
			isParent, err := helper.MOMIDHasParent(rlog, tx, momid, mt.ID)
			if err != nil {
				return model.ErrorToInternalServerErrorResponse(err)
			}
			if !isParent {
				return &model.Response{
					Status: fiber.StatusForbidden,
					Response: api.Error{
						Error: api.ErrorStrInsufficientCapabilities,
						ErrorDescription: fmt.Sprintf(
							"The provided token is neither a parent of the the token with "+
								" mom_id '%s' nor does it have the '%s' capability", momid,
							api.CapabilityNotifyAnyTokenRead.Name,
						),
					},
				}
			}
		}

		same, err := helper.CheckMytokensAreForSameUser(rlog, tx, momid, mt.ID)
		if err != nil {
			return model.ErrorToInternalServerErrorResponse(err)
		}
		if !same {
			return &model.Response{
				Status: fiber.StatusForbidden,
				Response: api.Error{
					Error: api.ErrorStrInvalidGrant,
					ErrorDescription: fmt.Sprintf(
						"The provided token cannot be used to obtain notifications for mom_id '%s'", momid,
					),
				},
			}
		}
	}
	return handleTokenInfoNotifications(rlog, tx, req, mt, clientMetadata)
}
