package tagging

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/internal/utils/mytokenutils"
)

// AddTagToMytokenRequest holds the request for adding a tag to a mytoken
type AddTagToMytokenRequest struct {
	api.AddTagToMytokenRequest
	Mytoken universalmytoken.UniversalMytoken `json:"mytoken" xml:"mytoken" form:"mytoken"`
	MomID   mtid.MOMID                        `json:"mom_id" xml:"mom_id" form:"mom_id"`
}

// RemoveTagFromMytokenRequest holds the request for removing a tag from a
// mytoken
type RemoveTagFromMytokenRequest struct {
	api.RemoveTagFromMytokenRequest
	Mytoken universalmytoken.UniversalMytoken `json:"mytoken" xml:"mytoken" form:"mytoken"`
	MomID   mtid.MOMID                        `json:"mom_id" xml:"mom_id" form:"mom_id"`
}

// HandleAddTagToMytoken handles requests to add a tag to a mytoken
func HandleAddTagToMytoken(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	var req AddTagToMytokenRequest
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.BadRequestErrorResponse(errorfmt.Error(err)).Send(ctx)
	}
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &req.Mytoken, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	clientMetadata := ctxutils.ClientMetaData(ctx)
	id, momMode, errRes := auth.ValidateCapabilityWithMomMode(
		rlog, api.CapabilityTokeninfoTags,
		api.CapabilityTagAnyToken, mt, req.MomID, clientMetadata,
	)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, nil, mt, clientMetadata)
	if errRes != nil {
		return errRes.Send(ctx)
	}

	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err := mytokenrepo.AddTag(
				rlog, tx, id.MomID(), req.Tag, req.IncludeChildren,
			); err != nil {
				return err
			}
			var rollback bool
			event := api.EventTagAddedToken
			if momMode {
				event = api.EventTagAddedTokenOther
			}
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *ctxutils.ClientMetaData(ctx),
				event, "", usedRestriction, req.Mytoken.JWT,
				req.Mytoken.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	)
	if err != nil && res == nil {
		res = model.ErrorToInternalServerErrorResponse(err)
	}
	if res == nil {
		res = &model.Response{Status: fiber.StatusNoContent}
	}
	return res.Send(ctx)
}

// HandleRemoveTagFromMytoken handles requests to remove a tag from a mytoken
func HandleRemoveTagFromMytoken(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	var req RemoveTagFromMytokenRequest
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.BadRequestErrorResponse(errorfmt.Error(err)).Send(ctx)
	}
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &req.Mytoken, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	clientMetadata := ctxutils.ClientMetaData(ctx)
	id, momMode, errRes := auth.ValidateCapabilityWithMomMode(
		rlog, api.CapabilityTokeninfoTags,
		api.CapabilityTagAnyToken, mt, req.MomID, clientMetadata,
	)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, nil, mt, clientMetadata)
	if errRes != nil {
		return errRes.Send(ctx)
	}

	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err := mytokenrepo.RemoveTag(
				rlog, tx, id.MomID(), req.Tag,
			); err != nil {
				return err
			}
			var rollback bool
			event := api.EventTagRemovedToken
			if momMode {
				event = api.EventTagRemovedTokenOther
			}
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *ctxutils.ClientMetaData(ctx),
				event, "", usedRestriction, req.Mytoken.JWT,
				req.Mytoken.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	)
	if err != nil && res == nil {
		res = model.ErrorToInternalServerErrorResponse(err)
	}
	if res == nil {
		res = &model.Response{Status: fiber.StatusNoContent}
	}
	return res.Send(ctx)
}
