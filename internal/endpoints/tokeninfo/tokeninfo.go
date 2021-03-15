package tokeninfo

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"

	dbhelper "github.com/oidc-mytoken/server/internal/db/dbrepo/supertokenrepo/supertokenrepohelper"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/ctxUtils"
	model2 "github.com/oidc-mytoken/server/pkg/model"
	supertoken "github.com/oidc-mytoken/server/shared/supertoken/pkg"
)

func HandleTokenInfo(ctx *fiber.Ctx) error {
	var req pkg.TokenInfoRequest
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	st, errRes := testSuperToken(ctx, &req)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	clientMetadata := ctxUtils.ClientMetaData(ctx)
	switch req.Action {
	case model2.TokeninfoActionIntrospect:
		return handleTokenInfoIntrospect(st, clientMetadata).Send(ctx)
	case model2.TokeninfoActionEventHistory:
		return handleTokenInfoHistory(st, clientMetadata).Send(ctx)
	case model2.TokeninfoActionSubtokenTree:
		return handleTokenInfoTree(st, clientMetadata).Send(ctx)
	case model2.TokeninfoActionListSuperTokens:
		return handleTokenInfoList(st, clientMetadata).Send(ctx)
	default:
		return model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model2.BadRequestError(fmt.Sprintf("unknown action '%s'", req.Action.String())),
		}.Send(ctx)
	}
}

func testSuperToken(ctx *fiber.Ctx, req *pkg.TokenInfoRequest) (*supertoken.SuperToken, *model.Response) {
	if req.SuperToken == "" {
		if t := ctxUtils.GetSuperToken(ctx); t != nil {
			req.SuperToken = *t
		} else {
			return nil, &model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: model2.InvalidTokenError("no super token found in request"),
			}
		}
	}

	st, err := supertoken.ParseJWT(string(req.SuperToken))
	if err != nil {
		return nil, &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model2.InvalidTokenError(err.Error()),
		}
	}

	revoked, dbErr := dbhelper.CheckTokenRevoked(st.ID)
	if dbErr != nil {
		return nil, model.ErrorToInternalServerErrorResponse(dbErr)
	}
	if revoked {
		return nil, &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model2.InvalidTokenError(""),
		}
	}
	return st, nil
}
