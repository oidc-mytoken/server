package tokeninfo

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	pkgModel "github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/mytoken/capabilities"
	eventService "github.com/oidc-mytoken/server/shared/mytoken/event"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
)

func handleTokenInfoIntrospect(st *mytoken.Mytoken, clientMetadata *model.ClientMetaData) model.Response {
	// If we call this function it means the token is valid.

	if !st.Capabilities.Has(capabilities.CapabilityTokeninfoIntrospect) {
		return model.Response{
			Status:   fiber.StatusForbidden,
			Response: pkgModel.APIErrorInsufficientCapabilities,
		}
	}
	var usedToken mytoken.UsedMytoken
	if err := db.RunWithinTransaction(nil, func(tx *sqlx.Tx) error {
		tmp, err := st.ToUsedMytoken(tx)
		usedToken = *tmp
		if err != nil {
			return err
		}
		return eventService.LogEvent(tx, eventService.MTEvent{
			Event: event.FromNumber(event.MTEventTokenInfoIntrospect, ""),
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
