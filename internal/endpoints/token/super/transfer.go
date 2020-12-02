package super

import (
	"encoding/json"

	uuid "github.com/satori/go.uuid"

	"github.com/jmoiron/sqlx"

	"github.com/zachmann/mytoken/internal/supertoken/event"
	event2 "github.com/zachmann/mytoken/internal/supertoken/event/pkg"

	"github.com/zachmann/mytoken/internal/utils/dbUtils"

	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/utils"

	"github.com/zachmann/mytoken/internal/db"

	"github.com/gofiber/fiber/v2"
	"github.com/zachmann/mytoken/internal/endpoints/token/super/pkg"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/utils/ctxUtils"
)

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

	transferCode := utils.RandASCIIString(config.Get().Features.TransferCodes.Len)
	expiresIn := config.Get().Features.TransferCodes.ExpiresAfter
	var stid uuid.UUID
	if err := db.Transact(func(tx *sqlx.Tx) error {
		if err := tx.Get(&stid, `SELECT id FROM SuperTokens WHERE token=?`, superToken); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO TransferCodes (transfer_code, ST_id, expires_in, new_st) VALUES(?,?,?,0)`, transferCode, stid, expiresIn); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	if err := event.LogEvent(&event2.Event{
		Type:    event2.STEventTransferCodeCreated,
		Comment: "from existing ST",
	}, stid, *ctxUtils.NetworkData(ctx)); err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	res := &model.Response{
		Status: fiber.StatusOK,
		Response: pkg.TransferCodeResponse{
			SuperTokenType: model.ResponseTypeTransferCode,
			TransferCode:   transferCode,
			ExpiresIn:      uint64(expiresIn),
		},
	}
	return res.Send(ctx)
}
