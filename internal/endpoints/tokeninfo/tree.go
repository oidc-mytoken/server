package tokeninfo

import (
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/shared/mytoken/event"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
)

func handleTokenInfoTree(mt *mytoken.Mytoken, clientMetadata *api.ClientMetaData) model.Response {
	// If we call this function it means the token is valid.

	if !mt.Capabilities.Has(api.CapabilityTokeninfoTree) {
		return model.Response{
			Status:   fiber.StatusForbidden,
			Response: api.ErrorInsufficientCapabilities,
		}
	}

	var usedRestriction *restrictions.Restriction
	if len(mt.Restrictions) > 0 {
		possibleRestrictions := mt.Restrictions.GetValidForOther(nil, clientMetadata.IP, mt.ID)
		if len(possibleRestrictions) == 0 {
			return model.Response{
				Status:   fiber.StatusForbidden,
				Response: api.ErrorUsageRestricted,
			}
		}
		usedRestriction = &possibleRestrictions[0]
	}

	var tokenTree tree.MytokenEntryTree
	if err := db.Transact(func(tx *sqlx.Tx) error {
		var err error
		tokenTree, err = tree.TokenSubTree(tx, mt.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if usedRestriction == nil {
			return nil
		}
		if err = usedRestriction.UsedOther(tx, mt.ID); err != nil {
			return err
		}
		return eventService.LogEvent(tx, eventService.MTEvent{
			Event: event.FromNumber(event.MTEventTokenInfoTree, ""),
			MTID:  mt.ID,
		}, *clientMetadata)
	}); err != nil {
		return *model.ErrorToInternalServerErrorResponse(err)
	}

	return model.Response{
		Status:   fiber.StatusOK,
		Response: pkg.NewTokeninfoTreeResponse(tokenTree),
	}
}
