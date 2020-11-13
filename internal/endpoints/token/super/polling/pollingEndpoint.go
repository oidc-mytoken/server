package polling

import (
	"encoding/json"
	"log"

	supertoken "github.com/zachmann/mytoken/internal/supertoken/pkg"

	"github.com/jmoiron/sqlx"

	"github.com/zachmann/mytoken/internal/db"

	"github.com/zachmann/mytoken/internal/db/dbModels"

	"github.com/gofiber/fiber/v2"
	response "github.com/zachmann/mytoken/internal/endpoints/token/super/pkg"
	"github.com/zachmann/mytoken/internal/model"
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
	log.Printf("Handle polling code '%s'", pollingCode)
	pollingCodeStatus, err := dbModels.CheckPollingCode(pollingCode)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if !pollingCodeStatus.Found {
		log.Printf("Polling code '%s' not known", pollingCode)
		return model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.APIErrorBadPollingCode,
		}
	}
	if pollingCodeStatus.Expired {
		log.Printf("Polling code '%s' expired", pollingCode)
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
		log.Printf("Retrieved token '%s' for polling code '%s' from db", token, pollingCode)
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
	log.Printf("The JWT was parsed as '%+v'", st)
	return model.Response{
		Status:   fiber.StatusOK,
		Response: st.ToSuperTokenResponse(token),
	}
}
