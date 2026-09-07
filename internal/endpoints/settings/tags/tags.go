package tags

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/service/tag"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleGet handles GET requests to the tags endpoint
func HandleGet(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle list tags request")
	var reqMytoken universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &reqMytoken, ctx)
	if errRes != nil {
		return errRes
	}
	return tag.Service.List(rlog, mt, reqMytoken, ctxutils.ClientMetaData(ctx))
}

// HandlePut handles PUT requests to the tags endpoint, updating a tag
func HandlePut(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle update tag request")
	var req api.TagInfo
	tagName := ctx.Params("tag")
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	if req.Tag == "" && req.Color == "" {
		return model.BadRequestErrorResponse("no supported request parameter given")
	}
	var reqMytoken universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &reqMytoken, ctx)
	if errRes != nil {
		return errRes
	}
	res := tag.Service.Update(rlog, mt, reqMytoken, ctxutils.ClientMetaData(ctx), tagName, req)
	if res == nil {
		return &model.Response{Status: http.StatusNoContent}
	}
	return res
}

// HandleDelete handles DELETE requests to the tags endpoint, deleting a tag
func HandleDelete(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle delete tag request")
	tagName := ctx.Params("tag")
	var reqMytoken universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &reqMytoken, ctx)
	if errRes != nil {
		return errRes
	}
	res := tag.Service.Delete(rlog, mt, reqMytoken, ctxutils.ClientMetaData(ctx), tagName)
	if res == nil {
		return &model.Response{Status: http.StatusNoContent}
	}
	return res
}

// HandlePost handles POST requests to the tags endpoint, creating a new tag
func HandlePost(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle create tag request")
	tagName := ctx.Params("tag")
	var req api.TagInfo
	body := ctx.Body()
	if len(body) > 0 {
		if err := ctx.BodyParser(&req); err != nil {
			return model.ErrorToBadRequestErrorResponse(err)
		}
	}
	var reqMytoken universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &reqMytoken, ctx)
	if errRes != nil {
		return errRes
	}
	res := tag.Service.Create(rlog, mt, reqMytoken, ctxutils.ClientMetaData(ctx), tagName, req.Color)
	if res == nil {
		return &model.Response{Status: http.StatusNoContent}
	}
	return res
}
