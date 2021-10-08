package grants

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/grantrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/settings"
	request "github.com/oidc-mytoken/server/internal/endpoints/settings/grants/pkg"
	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	serverModel "github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/shared/model"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/mytoken/universalmytoken"
)

// HandleListGrants handles GET requests to the grants endpoints and returns a list of enabled/disabled grant types for
// the user
func HandleListGrants(ctx *fiber.Ctx) error {
	log.Debug("Handle get grant type request")
	var reqMytoken universalmytoken.UniversalMytoken

	return settings.HandleSettings(
		ctx, &reqMytoken, event.FromNumber(event.GrantsListed, ""), fiber.StatusOK,
		func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *serverModel.Response) {
			grants, err := grantrepo.Get(tx, mt.ID)
			if err != nil {
				return nil, serverModel.ErrorToInternalServerErrorResponse(err)
			}
			return &request.GrantTypeInfoResponse{
				GrantTypeInfoResponse: api.GrantTypeInfoResponse{
					GrantTypes: grants,
				},
			}, nil
		},
	)
}

// HandleEnableGrant handles requests to enable a grant type
func HandleEnableGrant(ctx *fiber.Ctx) error {
	log.Debug("Handle enable grant type request")
	return handleEditGrant(ctx, grantrepo.Enable, event.GrantEnabled, fiber.StatusCreated)
}

// HandleDisableGrant handles requests to disable a grant type
func HandleDisableGrant(ctx *fiber.Ctx) error {
	log.Debug("Handle disable grant type request")
	return handleEditGrant(ctx, grantrepo.Disable, event.GrantDisabled, fiber.StatusNoContent)
}

func handleEditGrant(
	ctx *fiber.Ctx, dbCallBack func(tx *sqlx.Tx, myid mtid.MTID, grant model.GrantType) error, evt, okStatus int,
) error {
	req := request.GrantTypeRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return serverModel.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	log.Trace("Parsed grant type request")

	return settings.HandleSettings(
		ctx, &req.Mytoken, event.FromNumber(evt, req.GrantType.String()), okStatus,
		func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *serverModel.Response) {
			if err := dbCallBack(tx, mt.ID, req.GrantType); err != nil {
				return nil, serverModel.ErrorToInternalServerErrorResponse(err)
			}
			return nil, nil
		},
	)
}
