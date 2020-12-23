package super

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	dbhelper "github.com/zachmann/mytoken/internal/db/dbrepo/supertokenrepo/supertokenrepohelper"
	"github.com/zachmann/mytoken/internal/db/dbrepo/supertokenrepo/transfercoderepo"
	"github.com/zachmann/mytoken/internal/endpoints/token/super/pkg"
	"github.com/zachmann/mytoken/internal/model"
	supertoken "github.com/zachmann/mytoken/internal/supertoken/pkg"
	"github.com/zachmann/mytoken/internal/utils"
	"github.com/zachmann/mytoken/internal/utils/ctxUtils"
)

// HandleCreateTransferCodeForExistingSuperToken handles request to create a transfer code for an existing supertoken
func HandleCreateTransferCodeForExistingSuperToken(ctx *fiber.Ctx) error {
	token := ctxUtils.GetAuthHeaderToken(ctx)
	if len(token) == 0 {
		var req pkg.CreateTransferCodeRequest
		if err := json.Unmarshal(ctx.Body(), &req); err != nil {
			return model.Response{
				Status:   fiber.StatusBadRequest,
				Response: model.BadRequestError(err.Error()),
			}.Send(ctx)
		}
		token = req.SuperToken
		if len(token) == 0 {
			return model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: model.BadRequestError("required parameter 'super_token' missing"),
			}.Send(ctx)
		}
	}

	var jwt string
	var tokenType model.ResponseType
	if utils.IsJWT(token) {
		jwt = token
		tokenType = model.ResponseTypeToken
	} else {
		tokenType = model.ResponseTypeShortToken
		shortToken := transfercoderepo.ParseShortToken(token)
		tmp, valid, err := shortToken.JWT(nil)
		jwt = tmp
		if !valid {
			return model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: model.InvalidTokenError("invalid token"),
			}.Send(ctx)
		}
		if err != nil {
			return model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: model.InvalidTokenError(err.Error()),
			}.Send(ctx)
		}
	}
	superToken, err := supertoken.ParseJWT(jwt)
	if err != nil {
		return model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.InvalidTokenError(err.Error()),
		}.Send(ctx)
	}

	revoked, dbErr := dbhelper.CheckTokenRevoked(superToken.ID)
	if dbErr != nil {
		return model.ErrorToInternalServerErrorResponse(dbErr).Send(ctx)
	}
	if revoked {
		return model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.InvalidTokenError("invalid token"),
		}.Send(ctx)
	}

	transferCode, expiresIn, err := supertoken.CreateTransferCode(superToken.ID, token, false, tokenType, *ctxUtils.ClientMetaData(ctx))
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	res := &model.Response{
		Status: fiber.StatusOK,
		Response: pkg.TransferCodeResponse{
			SuperTokenType: model.ResponseTypeTransferCode,
			TransferCode:   transferCode,
			ExpiresIn:      expiresIn,
		},
	}
	return res.Send(ctx)
}
