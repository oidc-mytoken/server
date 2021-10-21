package tokeninfo

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	eventService "github.com/oidc-mytoken/server/shared/mytoken/event"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
)

func handleTokenInfoIntrospect(
	rlog log.Ext1FieldLogger,
	_ pkg.TokenInfoRequest,
	mt *mytoken.Mytoken,
	clientMetadata *api.ClientMetaData,
) model.Response {
	// If we call this function it means the token is valid.
	if errRes := auth.RequireCapability(rlog, api.CapabilityTokeninfoIntrospect, mt); errRes != nil {
		return *errRes
	}

	var usedToken mytoken.UsedMytoken
	if err := db.RunWithinTransaction(
		rlog, nil, func(tx *sqlx.Tx) error {
			tmp, err := mt.ToUsedMytoken(rlog, tx)
			if err != nil {
				return err
			}
			usedToken = *tmp
			return eventService.LogEvent(
				rlog, tx, eventService.MTEvent{
					Event: event.FromNumber(event.TokenInfoIntrospect, ""),
					MTID:  mt.ID,
				}, *clientMetadata,
			)
		},
	); err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
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
