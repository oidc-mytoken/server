package tokeninfo

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"

	dbhelper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/ctxUtils"
	model2 "github.com/oidc-mytoken/server/shared/model"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
)

// HandleTokenInfo handles requests to the tokeninfo endpoint
func HandleTokenInfo(ctx *fiber.Ctx) error {
	var req pkg.TokenInfoRequest
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	mt, errRes := testMytoken(ctx, &req)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	clientMetadata := ctxUtils.ClientMetaData(ctx)
	switch req.Action {
	case model2.TokeninfoActionIntrospect:
		return handleTokenInfoIntrospect(req, mt, clientMetadata).Send(ctx)
	case model2.TokeninfoActionEventHistory:
		return handleTokenInfoHistory(req, mt, clientMetadata).Send(ctx)
	case model2.TokeninfoActionSubtokenTree:
		return handleTokenInfoTree(req, mt, clientMetadata).Send(ctx)
	case model2.TokeninfoActionListMytokens:
		return handleTokenInfoList(req, mt, clientMetadata).Send(ctx)
	default:
		return model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model2.BadRequestError(fmt.Sprintf("unknown action '%s'", req.Action.String())),
		}.Send(ctx)
	}
}

func testMytoken(ctx *fiber.Ctx, req *pkg.TokenInfoRequest) (*mytoken.Mytoken, *model.Response) {
	if req.Mytoken.JWT == "" {
		if t := ctxUtils.GetMytoken(ctx); t != nil {
			req.Mytoken = *t
		} else {
			return nil, &model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: model2.InvalidTokenError("no mytoken found in request"),
			}
		}
	}

	mt, err := mytoken.ParseJWT(req.Mytoken.JWT)
	if err != nil {
		return nil, &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model2.InvalidTokenError(err.Error()),
		}
	}

	revoked, dbErr := dbhelper.CheckTokenRevoked(nil, mt.ID, mt.SeqNo, mt.Rotation)
	if dbErr != nil {
		return nil, model.ErrorToInternalServerErrorResponse(dbErr)
	}
	if revoked {
		return nil, &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model2.InvalidTokenError(""),
		}
	}
	return mt, nil
}
