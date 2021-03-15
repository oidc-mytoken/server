package tokeninfo

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	pkgModel "github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/supertoken/capabilities"
	eventService "github.com/oidc-mytoken/server/shared/supertoken/event"
	event "github.com/oidc-mytoken/server/shared/supertoken/event/pkg"
	supertoken "github.com/oidc-mytoken/server/shared/supertoken/pkg"
)

func handleTokenInfoIntrospect(st *supertoken.SuperToken, clientMetadata *model.ClientMetaData) model.Response {
	// If we call this function it means the token is valid.

	if !st.Capabilities.Has(capabilities.CapabilityTokeninfoIntrospect) {
		return model.Response{
			Status:   fiber.StatusForbidden,
			Response: pkgModel.APIErrorInsufficientCapabilities,
		}
	}
	var usedToken supertoken.UsedSuperToken
	if err := db.RunWithinTransaction(nil, func(tx *sqlx.Tx) error {
		tmp, err := st.ToUsedSuperToken(tx)
		usedToken = *tmp
		if err != nil {
			return err
		}
		return eventService.LogEvent(tx, eventService.MTEvent{
			Event: event.FromNumber(event.STEventTokenInfoIntrospect, ""),
			MTID:  st.ID,
		}, *clientMetadata)
	}); err != nil {
		return *model.ErrorToInternalServerErrorResponse(err)
	}
	return model.Response{
		Status: fiber.StatusOK,
		Response: pkg.TokeninfoIntrospectResponse{
			Valid: true,
			Token: usedToken,
		},
	}
}
