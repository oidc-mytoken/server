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
	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/shared/model"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/mytoken/universalmytoken"
)

// HandleListGrants handles GET requests to the grants endpoints and returns a list of enabled/disabled grant types for
// the user
func HandleListGrants(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle get grant type request")
	var reqMytoken universalmytoken.UniversalMytoken

	return settings.HandleSettings(
		ctx, &reqMytoken, event.FromNumber(event.GrantsListed, ""), fiber.StatusOK,
		func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *serverModel.Response) {
			grants, err := grantrepo.Get(rlog, tx, mt.ID)
			if err != nil {
				return nil, serverModel.ErrorToInternalServerErrorResponse(err)
			}
			return &request.GrantTypeInfoResponse{
				GrantTypeInfoResponse: api.GrantTypeInfoResponse{
					GrantTypes: grants,
				},
			}, nil
		}, false,
	)
}

// HandleEnableGrant handles requests to enable a grant type
func HandleEnableGrant(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle enable grant type request")
	return handleEditGrant(rlog, ctx, grantrepo.Enable, event.GrantEnabled, fiber.StatusCreated)
}

// HandleDisableGrant handles requests to disable a grant type
func HandleDisableGrant(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle disable grant type request")
	return handleEditGrant(rlog, ctx, grantrepo.Disable, event.GrantDisabled, fiber.StatusNoContent)
}

func handleEditGrant(
	rlog log.Ext1FieldLogger, ctx *fiber.Ctx,
	dbCallBack func(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myid mtid.MTID, grant model.GrantType) error,
	evt, okStatus int,
) error {
	req := request.GrantTypeRequest{GrantType: -1}
	if err := ctx.BodyParser(&req); err != nil {
		return serverModel.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	if !req.GrantType.Valid() {
		return serverModel.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError("no valid 'grant_type' found"),
		}.Send(ctx)
	}
	rlog.Trace("Parsed grant type request")

	return settings.HandleSettings(
		ctx, &req.Mytoken, event.FromNumber(evt, req.GrantType.String()), okStatus,
		func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *serverModel.Response) {
			if err := dbCallBack(rlog, tx, mt.ID, req.GrantType); err != nil {
				return nil, serverModel.ErrorToInternalServerErrorResponse(err)
			}
			return nil, nil
		}, false,
	)
}
