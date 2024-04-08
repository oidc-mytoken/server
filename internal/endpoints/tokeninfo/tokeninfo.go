package tokeninfo

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"

	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
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
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &req.Mytoken, ctx)
	if errRes != nil {
		return errRes
	}
	clientMetadata := ctxutils.ClientMetaData(ctx)
	switch req.Action {
	case model.TokeninfoActionIntrospect:
		return HandleTokenInfoIntrospect(rlog, nil, mt, req.Mytoken.OriginalTokenType, clientMetadata)
	case model.TokeninfoActionNotifications:
		return HandleTokenInfoNotifications(rlog, nil, &req, mt, clientMetadata)
	case model.TokeninfoActionEventHistory:
		return HandleTokenInfoHistory(rlog, nil, &req, mt, clientMetadata)
	case model.TokeninfoActionSubtokenTree:
		return HandleTokenInfoSubtokens(rlog, nil, &req, mt, clientMetadata)
	case model.TokeninfoActionListMytokens:
		return HandleTokenInfoList(rlog, nil, &req, mt, clientMetadata)
	default:
		return model.BadRequestErrorResponse(fmt.Sprintf("unknown action '%s'", req.Action.String()))
	}
}

func makeTokenInfoResponse(rsp interface{}, tokenUpdate *response.MytokenResponse) *model.Response {
	var cake []*fiber.Cookie
	if tokenUpdate != nil {
		cake = []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)}
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: rsp,
		Cookies:  cake,
	}
}
