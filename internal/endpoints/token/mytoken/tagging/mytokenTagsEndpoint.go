package tagging

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/service/mytokentag"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
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
		return model.BadRequestErrorResponse(err.Error()).Send(ctx)
	}
	mt, errRes := auth.RequireMytoken(rlog, &req.Mytoken, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	clientMetadata := ctxutils.ClientMetaData(ctx)
	if !req.MomID.HashValid() {
		req.MomID = mt.ID.MomID()
	}
	res := mytokentag.Service.AddTag(
		rlog, mt, req.Mytoken, clientMetadata, string(req.Tag), req.MomID, req.IncludeChildren,
	)
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
		return model.BadRequestErrorResponse(err.Error()).Send(ctx)
	}
	mt, errRes := auth.RequireMytoken(rlog, &req.Mytoken, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	clientMetadata := ctxutils.ClientMetaData(ctx)
	if !req.MomID.HashValid() {
		req.MomID = mt.ID.MomID()
	}
	res := mytokentag.Service.RemoveTag(
		rlog, mt, req.Mytoken, clientMetadata, string(req.Tag), req.MomID,
	)
	if res == nil {
		res = &model.Response{Status: fiber.StatusNoContent}
	}
	return res.Send(ctx)
}
