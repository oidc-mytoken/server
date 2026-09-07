package tokeninfo

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/eventrepo"
	helper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	pkg2 "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/rotation"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

// History returns the event history for the token.
func (s *service) History(
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
			res, rollback = s.historyLogic(rlog, tx, req, mt, clientMetaData)
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

func (s *service) historyLogic(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData,
) (*model.Response, bool) {
	if len(req.MOMIDs) == 0 {
		if errRes := auth.RequireCapability(
			rlog, tx, api.CapabilityTokeninfoHistory, mt, clientMetaData,
		); errRes != nil {
			return errRes, true
		}
		return s.handleTokenInfoHistory(rlog, tx, req, mt, clientMetaData)
	}
	for _, momid := range req.MOMIDs {
		if !mt.Capabilities.Has(api.CapabilityHistoryAnyToken) {
			if momid == api.MOMIDValueThis || momid == api.MOMIDValueChildren {
				continue
			}
			if strings.HasPrefix(momid, api.MOMIDValueChildren+"@") {
				momid = momid[len(api.MOMIDValueChildren)+1:]
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
							api.CapabilityHistoryAnyToken.Name,
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
						"The provided token cannot be used to obtain history for mom_id '%s'", momid,
					),
				},
			}, true
		}
	}
	return s.handleTokenInfoHistory(rlog, tx, req, mt, clientMetaData)
}

func (s *service) handleTokenInfoHistory(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData,
) (*model.Response, bool) {
	usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, tx, mt, clientMetaData)
	if errRes != nil {
		return errRes, true
	}
	history, tokenUpdate, err := s.doTokenInfoHistory(rlog, tx, req, mt, clientMetaData, usedRestriction)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err), true
	}
	rsp := pkg.NewTokeninfoHistoryResponse(history, tokenUpdate)
	return s.makeTokenInfoResponse(rsp, tokenUpdate), false
}

func (s *service) doTokenInfoHistory(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData,
	usedRestriction *restrictions.Restriction,
) (history eventrepo.EventHistory, tokenUpdate *response.MytokenResponse, err error) {
	var ids []any
	if len(req.MOMIDs) > 0 {
		for _, id := range req.MOMIDs {
			switch id {
			case api.MOMIDValueThis:
				ids = append(ids, mt.ID)
			case api.MOMIDValueChildren:
				history, err = eventrepo.GetEventHistoryChildren(rlog, tx, history, mt.ID)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					return
				}
			default:
				if strings.HasPrefix(id, api.MOMIDValueChildren+"@") {
					history, err = eventrepo.GetEventHistoryChildren(
						rlog, tx, history, id[len(api.MOMIDValueChildren)+1:],
					)
					if err != nil && !errors.Is(err, sql.ErrNoRows) {
						return
					}
				} else {
					ids = append(ids, id)
				}
			}
		}
	} else {
		ids = append(ids, mt.ID)
	}
	history, err = eventrepo.GetEventHistory(rlog, tx, history, ids...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}
	if usedRestriction == nil {
		return
	}
	if err = usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
		return
	}
	tokenUpdate, err = rotation.RotateMytokenAfterOtherForResponse(
		rlog, tx, req.Mytoken.JWT, mt, *clientMetaData, req.Mytoken.OriginalTokenType,
	)
	if err != nil {
		return
	}
	ev := api.EventTokenInfoHistory
	if len(req.MOMIDs) > 0 {
		ev = api.EventTokenInfoHistoryOtherToken
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
