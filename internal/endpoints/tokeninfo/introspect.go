package tokeninfo

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/shared/mytoken/event"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
)

func handleTokenInfoIntrospect(req pkg.TokenInfoRequest, mt *mytoken.Mytoken, clientMetadata *api.ClientMetaData) model.Response {
	// If we call this function it means the token is valid.

	if !mt.Capabilities.Has(api.CapabilityTokeninfoIntrospect) {
		return model.Response{
			Status:   fiber.StatusForbidden,
			Response: api.ErrorInsufficientCapabilities,
		}
	}
	var usedToken mytoken.UsedMytoken
	if err := db.RunWithinTransaction(nil, func(tx *sqlx.Tx) error {
		tmp, err := mt.ToUsedMytoken(tx)
		usedToken = *tmp
		if err != nil {
			return err
		}
		return eventService.LogEvent(tx, eventService.MTEvent{
			Event: event.FromNumber(event.MTEventTokenInfoIntrospect, ""),
			MTID:  mt.ID,
		}, *clientMetadata)
	}); err != nil {
		return *model.ErrorToInternalServerErrorResponse(err)
	}
	return model.Response{
		Status: fiber.StatusOK,
		Response: pkg.TokeninfoIntrospectResponse{
			TokeninfoIntrospectResponse: api.TokeninfoIntrospectResponse{
				Valid: true,
			},
			Token: usedToken,
		},
	}
}
