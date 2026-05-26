package tokeninfo

import (
	"errors"
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

// Notifications returns notification settings for the token.
func (s *service) Notifications(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData, req *pkg.TokenInfoRequest,
) *model.Response {
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			var rollback bool
			res, rollback = s.notificationsLogic(rlog, tx, req, mt, clientMetaData)
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

func (s *service) notificationsLogic(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData,
) (*model.Response, bool) {
	rlog.Debug("Handle tokeninfo notifications request")
	if len(req.MOMIDs) == 0 {
		if errRes := auth.RequireCapability(
			rlog, tx, api.CapabilityTokeninfoNotify, mt, clientMetaData,
		); errRes != nil {
			return errRes, true
		}
		return s.handleTokenInfoNotifications(rlog, tx, req, mt, clientMetaData)
	}
	for _, momid := range req.MOMIDs {
		if !mt.Capabilities.Has(api.CapabilityNotifyAnyTokenRead) {
			if momid == api.MOMIDValueThis {
				continue
			}
			isParent, err := helper.MOMIDHasParent(rlog, tx, momid, mt.ID)
			if err != nil {
				return model.ErrorToInternalServerErrorResponse(err), true
			}
			if !isParent {
				return &model.Response{
					Status: fiber.StatusForbidden,
					Response: api.Error{
						Error: api.ErrorStrInsufficientCapabilities,
						ErrorDescription: fmt.Sprintf(
							"The provided token is neither a parent of the token with "+
								" mom_id '%s' nor does it have the '%s' capability", momid,
							api.CapabilityNotifyAnyTokenRead.Name,
						),
					},
				}, true
			}
		}

		same, err := helper.CheckMytokensAreForSameUser(rlog, tx, momid, mt.ID)
		if err != nil {
			return model.ErrorToInternalServerErrorResponse(err), true
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
			}, true
		}
	}
	return s.handleTokenInfoNotifications(rlog, tx, req, mt, clientMetaData)
}

func (s *service) handleTokenInfoNotifications(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData,
) (*model.Response, bool) {
	usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, tx, mt, clientMetaData)
	if errRes != nil {
		return errRes, true
	}
	res, err := s.doTokenInfoNotifications(rlog, tx, req, mt, clientMetaData, usedRestriction)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err), true
	}
	rsp := &model.Response{
		Status:   fiber.StatusOK,
		Response: res,
	}
	if res.TokenUpdate != nil {
		rsp.Cookies = []*fiber.Cookie{cookies.MytokenCookie(res.TokenUpdate.Mytoken)}
	}
	return rsp, false
}

func (s *service) doTokenInfoNotifications(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData,
	usedRestriction *restrictions.Restriction,
) (res pkg.TokeninfoNotificationsResponse, err error) {
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
				return
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
			return
		}
	}
	if usedRestriction == nil {
		return
	}
	if err = usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
		return
	}
	res.TokenUpdate, err = rotation.RotateMytokenAfterOtherForResponse(
		rlog, tx, req.Mytoken.JWT, mt, *clientMetaData, req.Mytoken.OriginalTokenType,
	)
	if err != nil {
		return
	}
	ev := api.EventTokenInfoNotifications
	if len(req.MOMIDs) > 0 {
		ev = api.EventTokenInfoNotificationsOtherToken
	}
	err = eventService.LogEvent(
		rlog, tx, pkg2.MTEvent{
			Event:          ev,
			MTID:           mt.ID,
			ClientMetaData: *clientMetaData,
		},
	)
	return
}
