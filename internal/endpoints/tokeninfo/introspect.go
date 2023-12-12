package tokeninfo

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	pkg2 "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

// HandleTokenInfoIntrospect handles token introspection
func HandleTokenInfoIntrospect(
	rlog log.Ext1FieldLogger,
	tx *sqlx.Tx,
	mt *mytoken.Mytoken,
	origionalTokenType model.ResponseType,
	clientMetadata *api.ClientMetaData,
) model.Response {
	// If we call this function it means the token is valid.
	if errRes := auth.RequireCapability(
		rlog, tx, api.CapabilityTokeninfoIntrospect, mt, clientMetadata,
	); errRes != nil {
		return *errRes
	}

	var usedToken mytoken.UsedMytoken
	if err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			tmp, err := mt.ToUsedMytoken(rlog, tx)
			if err != nil {
				return err
			}
			usedToken = *tmp
			return eventService.LogEvent(
				rlog, tx, pkg2.MTEvent{
					Event:          api.EventTokenInfoIntrospect,
					MTID:           mt.ID,
					ClientMetaData: *clientMetadata,
				},
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
				MOMID: mt.ID.Hash(),
			},
			Token:     usedToken,
			TokenType: origionalTokenType,
		},
	}
}
