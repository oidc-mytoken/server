package email

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/service/email"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleGet handles GET requests to the email settings endpoint
func HandleGet(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle get email info request")
	var reqMytoken universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &reqMytoken, ctx)
	if errRes != nil {
		return errRes
	}
	return email.Service.Get(rlog, mt, reqMytoken, ctxutils.ClientMetaData(ctx))
}

// HandlePut handles PUT requests to the email settings endpoint, i.e. it updates email settings
func HandlePut(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle update email settings request")
	var req api.UpdateMailSettingsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	if req.PreferHTMLMail == nil && req.EmailAddress == "" {
		return model.BadRequestErrorResponse("no request parameter given")
	}
	var reqMytoken universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireMytoken(rlog, &reqMytoken, ctx)
	if errRes != nil {
		return errRes
	}
	res := email.Service.Set(rlog, mt, reqMytoken, ctxutils.ClientMetaData(ctx), req)
	if res == nil {
		return &model.Response{Status: fiber.StatusNoContent}
	}
	return res
}
