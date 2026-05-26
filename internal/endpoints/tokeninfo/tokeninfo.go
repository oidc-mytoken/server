package tokeninfo

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/service/tokeninfo"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleTokenInfo handles requests to the tokeninfo endpoint
func HandleTokenInfo(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	var req pkg.TokenInfoRequest
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	mt, errRes := auth.RequireMytoken(rlog, &req.Mytoken, ctx)
	if errRes != nil {
		return errRes
	}
	clientMetadata := ctxutils.ClientMetaData(ctx)
	switch req.Action {
	case model.TokeninfoActionIntrospect:
		return tokeninfo.Service.Introspect(rlog, mt, req.Mytoken.OriginalTokenType, clientMetadata)
	case model.TokeninfoActionNotifications:
		return tokeninfo.Service.Notifications(rlog, mt, clientMetadata, &req)
	case model.TokeninfoActionEventHistory:
		return tokeninfo.Service.History(rlog, mt, clientMetadata, &req)
	case model.TokeninfoActionSubtokenTree:
		return tokeninfo.Service.Subtokens(rlog, mt, clientMetadata, &req)
	case model.TokeninfoActionListMytokens:
		return tokeninfo.Service.List(rlog, mt, clientMetadata, &req)
	default:
		return model.BadRequestErrorResponse(fmt.Sprintf("unknown action '%s'", req.Action.String()))
	}
}
