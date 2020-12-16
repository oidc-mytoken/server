package super

import (
	"encoding/json"

	supertoken "github.com/zachmann/mytoken/internal/supertoken/pkg"

	uuid "github.com/satori/go.uuid"

	"github.com/zachmann/mytoken/internal/utils/dbUtils"

	"github.com/zachmann/mytoken/internal/db"

	"github.com/gofiber/fiber/v2"

	"github.com/zachmann/mytoken/internal/endpoints/token/super/pkg"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/utils/ctxUtils"
)

// HandleCreateTransferCodeForExistingSuperToken handles request to create a transfer code for an existing supertoken
func HandleCreateTransferCodeForExistingSuperToken(ctx *fiber.Ctx) error {
	token := ctxUtils.GetAuthHeaderToken(ctx)
	if len(token) == 0 {
		var req pkg.CreateTransferCodeRequest
		if err := json.Unmarshal(ctx.Body(), &req); err != nil {
			res := &model.Response{
				Status:   fiber.StatusBadRequest,
				Response: model.BadRequestError(err.Error()),
			}
			return res.Send(ctx)
		}
		token = req.SuperToken
		if len(token) == 0 {
			res := &model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: model.BadRequestError("required parameter 'super_token' missing"),
			}
			return res.Send(ctx)
		}
	}

	superToken, revoked, dbErr := dbUtils.CheckTokenRevoked(token)
	if dbErr != nil {
		return model.ErrorToInternalServerErrorResponse(dbErr).Send(ctx)
	}
	if revoked {
		res := &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.InvalidTokenError("token not valid"),
		}
		return res.Send(ctx)
	}

	var stid uuid.UUID
	if err := db.DB().Get(&stid, `SELECT id FROM SuperTokens WHERE token=?`, superToken); err != nil {
		return err
	}
	transferCode, expiresIn, err := supertoken.CreateTransferCode(stid, *ctxUtils.ClientMetaData(ctx))
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
