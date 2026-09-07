package access

import (
	"github.com/gofiber/fiber/v2"

	request "github.com/oidc-mytoken/server/internal/endpoints/token/access/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/service/access"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleAccessTokenEndpoint handles request on the access token endpoint
func HandleAccessTokenEndpoint(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle access token request")
	req := request.NewAccessTokenRequest()
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	rlog.Trace("Parsed access token request")
	if req.Mytoken.JWT == "" {
		req.Mytoken = req.RefreshToken
	}

	if errRes := auth.RequireGrantType(rlog, model.GrantTypeMytoken, req.GrantType); errRes != nil {
		return errRes
	}
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &req.Mytoken, ctx)
	if errRes != nil {
		return errRes
	}

	return access.Service.CreateAccessToken(
		rlog, mt, ctxutils.ClientMetaData(ctx), req,
	)
}
