package super

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	"github.com/zachmann/mytoken/internal/db/dbrepo/supertokenrepo/transfercoderepo"
	"github.com/zachmann/mytoken/internal/endpoints/token/super/pkg"
	"github.com/zachmann/mytoken/internal/model"
	dbhelper "github.com/zachmann/mytoken/internal/server/db/dbrepo/supertokenrepo/supertokenrepohelper"
	supertoken "github.com/zachmann/mytoken/internal/supertoken/pkg"
	"github.com/zachmann/mytoken/internal/utils/ctxUtils"
	pkgModel "github.com/zachmann/mytoken/pkg/model"
	"github.com/zachmann/mytoken/shared/utils"
)

// HandleCreateTransferCodeForExistingSuperToken handles request to create a transfer code for an existing supertoken
func HandleCreateTransferCodeForExistingSuperToken(ctx *fiber.Ctx) error {
	token := ctxUtils.GetAuthHeaderToken(ctx)
	if len(token) == 0 {
		var req pkg.CreateTransferCodeRequest
		if err := json.Unmarshal(ctx.Body(), &req); err != nil {
			return model.Response{
				Status:   fiber.StatusBadRequest,
				Response: pkgModel.BadRequestError(err.Error()),
			}.Send(ctx)
		}
		token = req.SuperToken
		if len(token) == 0 {
			return model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: pkgModel.BadRequestError("required parameter 'super_token' missing"),
			}.Send(ctx)
		}
	}

	var jwt string
	var tokenType pkgModel.ResponseType
	if utils.IsJWT(token) {
		jwt = token
		tokenType = pkgModel.ResponseTypeToken
	} else {
		tokenType = pkgModel.ResponseTypeShortToken
		shortToken := transfercoderepo.ParseShortToken(token)
		tmp, valid, err := shortToken.JWT(nil)
		jwt = tmp
		if !valid {
			return model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: pkgModel.InvalidTokenError("invalid token"),
			}.Send(ctx)
		}
		if err != nil {
			return model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: pkgModel.InvalidTokenError(err.Error()),
			}.Send(ctx)
		}
	}
	superToken, err := supertoken.ParseJWT(jwt)
	if err != nil {
		return model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: pkgModel.InvalidTokenError(err.Error()),
		}.Send(ctx)
	}

	revoked, dbErr := dbhelper.CheckTokenRevoked(superToken.ID)
	if dbErr != nil {
		return model.ErrorToInternalServerErrorResponse(dbErr).Send(ctx)
	}
	if revoked {
		return model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: pkgModel.InvalidTokenError("invalid token"),
		}.Send(ctx)
	}

	transferCode, expiresIn, err := supertoken.CreateTransferCode(superToken.ID, token, false, tokenType, *ctxUtils.ClientMetaData(ctx))
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	res := &model.Response{
		Status: fiber.StatusOK,
		Response: pkg.TransferCodeResponse{
			SuperTokenType: pkgModel.ResponseTypeTransferCode,
			TransferCode:   transferCode,
			ExpiresIn:      expiresIn,
		},
	}
	return res.Send(ctx)
}
