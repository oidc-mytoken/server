// Package revoke provides a service layer for token revocation operations.
package revoke

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	helper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/model"
	mytokenPkg "github.com/oidc-mytoken/server/internal/mytoken"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	"github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

// Service is the revoke service singleton.
var Service = &service{}

type service struct{}

// ByMOMID revokes a mytoken identified by its MOMID.
func (s *service) ByMOMID(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData, momID string, recursive bool,
) *model.Response {
	if !config.Get().Features.TokenRevocation.Enabled {
		return model.BadRequestErrorResponse("revocation is disabled")
	}
	if momID == "" {
		return model.BadRequestErrorResponse("mom_id is required")
	}
	var res *model.Response
	if err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			isParent, err := helper.MOMIDHasParent(rlog, tx, momID, mt.ID)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if !isParent && !mt.Capabilities.Has(api.CapabilityRevokeAnyToken) {
				res = &model.Response{
					Status: fiber.StatusForbidden,
					Response: api.Error{
						Error:            api.ErrorStrInsufficientCapabilities,
						ErrorDescription: "The provided token is neither a parent of the token to be revoked nor does it have the 'revoke_any_token' capability",
					},
				}
				return errors.New("rollback")
			}
			same, err := helper.CheckMytokensAreForSameUser(rlog, tx, momID, mt.ID)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if !same {
				res = &model.Response{
					Status: fiber.StatusForbidden,
					Response: api.Error{
						Error:            api.ErrorStrInvalidGrant,
						ErrorDescription: "The provided token cannot be used to revoke this mom_id",
					},
				}
				return errors.New("rollback")
			}
			if momID == mt.ID.Hash() {
				res = &model.Response{
					Status: fiber.StatusBadRequest,
					Response: api.Error{
						Error:            api.ErrorStrInvalidRequest,
						ErrorDescription: "A token cannot be revoked by its own mom_id. Use the token itself instead.",
					},
				}
				return errors.New("rollback")
			}
			if err = helper.RevokeMT(rlog, tx, momID, recursive); err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if err = eventService.LogEvent(
				rlog, tx, pkg.MTEvent{
					Event:          api.EventRevokedOtherToken,
					MTID:           mt.ID,
					Comment:        fmt.Sprintf("mom_id: %s", momID),
					ClientMetaData: *clientMetaData,
				},
			); err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
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

// RevokeSelf revokes the token itself.
func (s *service) RevokeSelf(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData,
) *model.Response {
	if !config.Get().Features.TokenRevocation.Enabled {
		return model.BadRequestErrorResponse("revocation is disabled")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			errRes := mytokenPkg.RevokeMytoken(rlog, tx, mt.ID, umt.JWT, false, mt.OIDCIssuer)
			if errRes != nil {
				res = errRes
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
