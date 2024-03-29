package tokeninfo

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/eventrepo"
	helper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	event "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/rotation"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

func doTokenInfoHistory(
	rlog log.Ext1FieldLogger, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken, clientMetadata *api.ClientMetaData,
	usedRestriction *restrictions.Restriction,
) (history eventrepo.EventHistory, tokenUpdate *response.MytokenResponse, err error) {
	err = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			var ids []any
			if len(req.MOMIDs) > 0 {
				for _, id := range req.MOMIDs {
					switch id {
					case api.MOMIDValueThis:
						ids = append(ids, mt.ID)
					case api.MOMIDValueChildren:
						history, err = eventrepo.GetEventHistoryChildren(rlog, tx, history, mt.ID)
						if err != nil && !errors.Is(err, sql.ErrNoRows) {
							return err
						}
					default:
						if strings.HasPrefix(id, api.MOMIDValueChildren+"@") {
							history, err = eventrepo.GetEventHistoryChildren(
								rlog, tx, history, id[len(api.MOMIDValueChildren)+1:],
							)
							if err != nil && !errors.Is(err, sql.ErrNoRows) {
								return err
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
				return err
			}
			if usedRestriction == nil {
				return nil
			}
			if err = usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
				return err
			}
			tokenUpdate, err = rotation.RotateMytokenAfterOtherForResponse(
				rlog, tx, req.Mytoken.JWT, mt, *clientMetadata, req.Mytoken.OriginalTokenType,
			)
			if err != nil {
				return err
			}
			ev := event.FromNumber(event.TokenInfoHistory, "")
			if len(req.MOMIDs) > 0 {
				ev = event.FromNumber(event.TokenInfoHistoryOtherToken, "")
			}
			return eventService.LogEvent(
				rlog, tx, eventService.MTEvent{
					Event: ev,
					MTID:  mt.ID,
				}, *clientMetadata,
			)
		},
	)
	return
}

func handleTokenInfoHistory(
	rlog log.Ext1FieldLogger, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken, clientMetadata *api.ClientMetaData,
) model.Response {
	usedRestriction, errRes := auth.RequireUsableRestrictionOther(
		rlog, nil, mt, clientMetadata.IP, nil, nil,
	)
	if errRes != nil {
		return *errRes
	}
	history, tokenUpdate, err := doTokenInfoHistory(rlog, req, mt, clientMetadata, usedRestriction)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return *model.ErrorToInternalServerErrorResponse(err)
	}
	rsp := pkg.NewTokeninfoHistoryResponse(history, tokenUpdate)
	return makeTokenInfoResponse(rsp, tokenUpdate)
}

// HandleTokenInfoHistory handles a tokeninfo history request
func HandleTokenInfoHistory(
	rlog log.Ext1FieldLogger, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken, clientMetadata *api.ClientMetaData,
) model.Response {
	// If we call this function it means the token is valid.

	if len(req.MOMIDs) == 0 {
		if errRes := auth.RequireCapability(rlog, api.CapabilityTokeninfoHistory, mt); errRes != nil {
			return *errRes
		}
		return handleTokenInfoHistory(rlog, req, mt, clientMetadata)
	}
	if !mt.Capabilities.Has(api.CapabilityHistoryAnyToken) {
		for _, momid := range req.MOMIDs {
			if momid == api.MOMIDValueThis || momid == api.MOMIDValueChildren {
				continue
			}
			if strings.HasPrefix(momid, api.MOMIDValueChildren+"@") {
				momid = momid[len(api.MOMIDValueChildren)+1:]
			}
			isParent, err := helper.MOMIDHasParent(rlog, nil, momid, mt.ID)
			if err != nil {
				return *model.ErrorToInternalServerErrorResponse(err)
			}
			if !isParent {
				return model.Response{
					Status: fiber.StatusForbidden,
					Response: api.Error{
						Error: api.ErrorStrInsufficientCapabilities,
						ErrorDescription: fmt.Sprintf(
							"The provided token is neither a parent of the the token with "+
								" mom_id '%s' nor does it have the '%s' capability", momid,
							api.CapabilityHistoryAnyToken.Name,
						),
					},
				}
			}

			same, err := helper.CheckMytokensAreForSameUser(rlog, nil, momid, mt.ID)
			if err != nil {
				return *model.ErrorToInternalServerErrorResponse(err)
			}
			if !same {
				return model.Response{
					Status: fiber.StatusForbidden,
					Response: api.Error{
						Error: api.ErrorStrInvalidGrant,
						ErrorDescription: fmt.Sprintf(
							"The provided token cannot be used to obtain history for mom_id '%s'", momid,
						),
					},
				}
			}
		}
	}
	return handleTokenInfoHistory(rlog, req, mt, clientMetadata)
}
