package tokeninfo

import (
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/eventrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	pkgModel "github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/mytoken/capabilities"
	eventService "github.com/oidc-mytoken/server/shared/mytoken/event"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
)

func handleTokenInfoHistory(st *mytoken.Mytoken, clientMetadata *model.ClientMetaData) model.Response {
	// If we call this function it means the token is valid.

	if !st.Capabilities.Has(capabilities.CapabilityTokeninfoHistory) {
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

	var history eventrepo.EventHistory
	if err := db.Transact(func(tx *sqlx.Tx) error {
		var err error
		history, err = eventrepo.GetEventHistory(tx, st.ID)
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
			Event: event.FromNumber(event.MTEventTokenInfoHistory, ""),
			MTID:  st.ID,
		}, *clientMetadata)
	}); err != nil {
		return *model.ErrorToInternalServerErrorResponse(err)
	}
	return model.Response{
		Status:   fiber.StatusOK,
		Response: pkg.NewTokeninfoHistoryResponse(history),
	}
}
