package tokeninfo

import (
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/supertokenrepo/tree"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	pkgModel "github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/supertoken/capabilities"
	eventService "github.com/oidc-mytoken/server/shared/supertoken/event"
	event "github.com/oidc-mytoken/server/shared/supertoken/event/pkg"
	supertoken "github.com/oidc-mytoken/server/shared/supertoken/pkg"
	"github.com/oidc-mytoken/server/shared/supertoken/restrictions"
)

func handleTokenInfoTree(st *supertoken.SuperToken, clientMetadata *model.ClientMetaData) model.Response {
	// If we call this function it means the token is valid.

	if !st.Capabilities.Has(capabilities.CapabilityTokeninfoTree) {
		return model.Response{
			Status:   fiber.StatusForbidden,
			Response: pkgModel.APIErrorInsufficientCapabilities,
		}
	}

	var usedRestriction *restrictions.Restriction
	if len(st.Restrictions) > 0 {
		possibleRestrictions := st.Restrictions.GetValidForOther(nil, clientMetadata.IP, st.ID)
		if len(possibleRestrictions) == 0 {
			return model.Response{
				Status:   fiber.StatusForbidden,
				Response: pkgModel.APIErrorUsageRestricted,
			}
		}
		usedRestriction = &possibleRestrictions[0]
	}

	var tokenTree tree.SuperTokenEntryTree
	if err := db.Transact(func(tx *sqlx.Tx) error {
		var err error
		tokenTree, err = tree.TokenSubTree(tx, st.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if usedRestriction == nil {
			return nil
		}
		if err = usedRestriction.UsedOther(tx, st.ID); err != nil {
			return err
		}
		return eventService.LogEvent(tx, eventService.MTEvent{
			Event: event.FromNumber(event.STEventTokenInfoTree, ""),
			MTID:  st.ID,
		}, *clientMetadata)
	}); err != nil {
		return *model.ErrorToInternalServerErrorResponse(err)
	}

	return model.Response{
		Status:   fiber.StatusOK,
		Response: pkg.NewTokeninfoTreeResponse(tokenTree),
	}
}
