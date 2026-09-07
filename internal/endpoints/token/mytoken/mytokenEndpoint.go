package mytoken

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/config"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/polling"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken2 "github.com/oidc-mytoken/server/internal/mytoken"
	"github.com/oidc-mytoken/server/internal/oidc/authcode"
	provider2 "github.com/oidc-mytoken/server/internal/oidc/provider"
	"github.com/oidc-mytoken/server/internal/service/mytoken"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleMytokenEndpoint handles requests on the mytoken endpoint
func HandleMytokenEndpoint(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	grantType, err := ctxutils.GetGrantType(ctx)
	if err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	rlog.WithField("grant_type", grantType.String()).Trace("Received mytoken request")
	switch grantType {
	case model.GrantTypeMytoken:
		return handleMytokenFromMytoken(ctx)
	case model.GrantTypeOIDCFlow:
		return handleOIDCFlow(ctx)
	case model.GrantTypePollingCode:
		if config.Get().Features.Polling.Enabled {
			return polling.HandlePollingCode(ctx)
		}
	case model.GrantTypeTransferCode:
		if config.Get().Features.TransferCodes.Enabled {
			return mytoken2.HandleMytokenFromTransferCode(ctx)
		}
	}
	return &model.Response{
		Status:   fiber.StatusBadRequest,
		Response: api.ErrorUnsupportedGrantType,
	}
}

func handleMytokenFromMytoken(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle mytoken from mytoken")
	req := response.NewMytokenRequest()
	if err := errors.WithStack(json.Unmarshal(ctx.Body(), &req)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	req.GrantType = model.GrantTypeMytoken
	mt, errRes := auth.RequireMytoken(rlog, &req.Mytoken, ctx)
	if errRes != nil {
		return errRes
	}
	return mytoken.Service.CreateFromMytoken(rlog, mt, req.Mytoken, ctxutils.ClientMetaData(ctx), req)
}

func handleOIDCFlow(ctx *fiber.Ctx) *model.Response {
	req := response.NewOIDCFlowRequest()
	if err := json.Unmarshal(ctx.Body(), req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	if p := provider2.GetProvider(req.Issuer); p == nil {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}
	}
	if len(req.Capabilities.Capabilities) == 0 {
		req.Capabilities.Capabilities = api.DefaultCapabilities
	}
	switch req.OIDCFlow {
	case model.OIDCFlowAuthorizationCode:
		authCodeReq := &response.AuthCodeFlowRequest{OIDCFlowRequest: *req}
		if err := json.Unmarshal(ctx.Body(), authCodeReq); err != nil {
			return model.ErrorToBadRequestErrorResponse(err)
		}
		return authcode.StartAuthCodeFlow(ctx, authCodeReq)
	default:
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnsupportedOIDCFlow,
		}
	}
}
