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
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleListGrants handles GET requests to the grants endpoints and returns a list of enabled/disabled grant types for
// the user
func HandleListGrants(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle get grant type request")
	var reqMytoken universalmytoken.UniversalMytoken

	return settings.HandleSettingsHelper(
		ctx, &reqMytoken, api.CapabilityGrantsRead, &api.EventGrantsListed, "", fiber.StatusOK,
		func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *model.Response) {
			grants, err := grantrepo.Get(rlog, tx, mt.ID)
			if err != nil {
				return nil, model.ErrorToInternalServerErrorResponse(err)
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
	return handleEditGrant(rlog, ctx, grantrepo.Enable, api.EventGrantEnabled, fiber.StatusCreated)
}

// HandleDisableGrant handles requests to disable a grant type
func HandleDisableGrant(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle disable grant type request")
	return handleEditGrant(rlog, ctx, grantrepo.Disable, api.EventGrantDisabled, fiber.StatusNoContent)
}

func handleEditGrant(
	rlog log.Ext1FieldLogger, ctx *fiber.Ctx,
	dbCallBack func(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myid mtid.MTID, grant model.GrantType) error,
	evt api.Event, okStatus int,
) error {
	req := request.GrantTypeRequest{GrantType: -1}
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	if !req.GrantType.Valid() {
		return model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError("no valid 'grant_type' found"),
		}.Send(ctx)
	}
	rlog.Trace("Parsed grant type request")

	return settings.HandleSettingsHelper(
		ctx, &req.Mytoken, api.CapabilityGrants, &evt, req.GrantType.String(), okStatus,
		func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *model.Response) {
			if err := dbCallBack(rlog, tx, mt.ID, req.GrantType); err != nil {
				return nil, model.ErrorToInternalServerErrorResponse(err)
			}
			return nil, nil
		}, false,
	)
}
