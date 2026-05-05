package tags

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/tagrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/settings"
	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/internal/utils/mytokenutils"
)

// TagListingResponse is the response type that lists tags of a user
type TagListingResponse struct {
	api.TagListingResponse
	TokenUpdate *my.MytokenResponse `json:"token_update,omitempty"`
}

// SetTokenUpdate implements the pkg.TokenUpdatableResponse interface
func (res *TagListingResponse) SetTokenUpdate(tokenUpdate *my.MytokenResponse) {
	res.TokenUpdate = tokenUpdate
}

// HandleGet handles GET requests to the tags endpoint
func HandleGet(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle list tags request")
	var reqMytoken universalmytoken.UniversalMytoken

	return settings.HandleSettingsHelper(
		ctx, nil, &reqMytoken, api.CapabilityTagsRead,
		&api.EventTagsListed, "", fiber.StatusOK,
		func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *model.Response) {
			tags, err := tagrepo.ListTags(rlog, tx, mt.ID)
			if err != nil {
				return nil, model.ErrorToInternalServerErrorResponse(err)
			}
			return &TagListingResponse{
				TagListingResponse: api.TagListingResponse{
					Tags: tags,
				},
			}, nil
		}, false,
	)
}

// HandlePut handles PUT requests to the tags endpoint, updating a tag
func HandlePut(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle update tag request")
	var req api.TagInfo
	tag := ctx.Params("tag")
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	if req.Tag == "" && req.Color == "" {
		return model.BadRequestErrorResponse(
			"no supported request parameter given",
		)
	}
	var reqMytoken universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &reqMytoken, ctx)
	if errRes != nil {
		return errRes
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt, ctxutils.ClientMetaData(ctx), api.CapabilityTags,
	)
	if errRes != nil {
		return errRes
	}
	var res *model.Response
	clientMetaData := ctxutils.ClientMetaData(ctx)
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err := tagrepo.UpdateTag(rlog, tx, tag, req, mt.ID); err != nil {
				return err
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				api.EventTagUpdated, tag, usedRestriction, reqMytoken.JWT,
				reqMytoken.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil && res == nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if res == nil {
		res = &model.Response{
			Status: http.StatusNoContent,
		}
	}
	return res
}

// HandleDelete handles DELETE requests to the tags endpoint, deleting a tag
func HandleDelete(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle delete tag request")
	tag := ctx.Params("tag")
	var reqMytoken universalmytoken.UniversalMytoken

	return settings.HandleSettingsHelper(
		ctx, nil, &reqMytoken, api.CapabilityTags,
		&api.EventTagDeleted, tag, fiber.StatusNoContent,
		func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *model.Response) {
			if err := tagrepo.DeleteTag(rlog, tx, tag, mt.ID); err != nil {
				return nil, model.ErrorToInternalServerErrorResponse(err)
			}
			return &my.OnlyTokenUpdateRes{}, nil
		}, false,
	)
}

// HandlePost handles POST requests to the tags endpoint, creating a new tag
func HandlePost(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle create tag request")
	tag := ctx.Params("tag")
	var req api.TagInfo
	body := ctx.Body()
	if len(body) > 0 {
		if err := ctx.BodyParser(&req); err != nil {
			return model.ErrorToBadRequestErrorResponse(err)
		}
	}
	var reqMytoken universalmytoken.UniversalMytoken

	return settings.HandleSettingsHelper(
		ctx, nil, &reqMytoken, api.CapabilityTags,
		&api.EventTagCreated, tag, fiber.StatusNoContent,
		func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *model.Response) {
			if err := tagrepo.CreateTag(rlog, tx, tag, req.Color, mt.ID); err != nil {
				return nil, model.ErrorToInternalServerErrorResponse(err)
			}
			return &my.OnlyTokenUpdateRes{}, nil
		}, false,
	)
}
