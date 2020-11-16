package polling

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/db/dbModels"
	response "github.com/zachmann/mytoken/internal/endpoints/token/super/pkg"
	"github.com/zachmann/mytoken/internal/model"
	supertoken "github.com/zachmann/mytoken/internal/supertoken/pkg"
)

func HandlePollingCode(ctx *fiber.Ctx) error {
	req := response.PollingCodeRequest{}
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		res := model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError(err.Error()),
		}
		return res.Send(ctx)
	}
	res := handlePollingCode(req)
	return res.Send(ctx)
}

func handlePollingCode(req response.PollingCodeRequest) model.Response {
	pollingCode := req.PollingCode
	log.WithField("polling_code", pollingCode).Debug("Handle polling code")
	pollingCodeStatus, err := dbModels.CheckPollingCode(pollingCode)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if !pollingCodeStatus.Found {
		log.WithField("polling_code", pollingCode).Debug("Polling code not known")
		return model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.APIErrorBadPollingCode,
		}
	}
	if pollingCodeStatus.Expired {
		log.WithField("polling_code", pollingCode).Debug("Polling code expired")
		return model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.APIErrorPollingCodeExpired,
		}
	}
	var token string
	if err := db.Transact(func(tx *sqlx.Tx) error {
		if err := tx.Get(&token, `SELECT token FROM TmpST_by_polling_code WHERE polling_code=? AND CURRENT_TIMESTAMP() <= polling_code_expires_at`, pollingCode); err != nil {
			return err
		}
		log.WithFields(log.Fields{"token": token, "polling_code": pollingCode}).Debug("Retrieved token for polling code from db")
		if token == "" {
			return nil
		}
		if _, err := tx.Exec(`DELETE FROM PollingCodes WHERE polling_code=?`, pollingCode); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if token == "" {
		return model.Response{
			Status:   fiber.StatusPreconditionRequired,
			Response: model.APIErrorAuthorizationPending,
		}
	}
	st, err := supertoken.ParseJWT(token)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	log.Tracef("The JWT was parsed as '%+v'", st)
	return model.Response{
		Status:   fiber.StatusOK,
		Response: st.ToSuperTokenResponse(token),
	}
}
