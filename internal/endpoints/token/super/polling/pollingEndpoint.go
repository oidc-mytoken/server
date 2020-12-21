package polling

import (
	"encoding/json"

	"github.com/zachmann/mytoken/internal/db/dbrepo/pollingcoderepo"
	"github.com/zachmann/mytoken/internal/oidc/authcode"

	"github.com/zachmann/mytoken/internal/utils/ctxUtils"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	response "github.com/zachmann/mytoken/internal/endpoints/token/super/pkg"
	"github.com/zachmann/mytoken/internal/model"
	supertoken "github.com/zachmann/mytoken/internal/supertoken/pkg"
)

// HandlePollingCode handles a request on the polling endpoint
func HandlePollingCode(ctx *fiber.Ctx) error {
	req := response.PollingCodeRequest{}
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	return handlePollingCode(req, *ctxUtils.ClientMetaData(ctx)).Send(ctx)
}

func handlePollingCode(req response.PollingCodeRequest, networkData model.ClientMetaData) *model.Response {
	pollingCode := req.PollingCode
	log.WithField("polling_code", pollingCode).Debug("Handle polling code")
	pollingCodeStatus, err := pollingcoderepo.CheckPollingCode(nil, pollingCode)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if !pollingCodeStatus.Found {
		log.WithField("polling_code", pollingCode).Debug("Polling code not known")
		return &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.APIErrorBadPollingCode,
		}
	}
	if pollingCodeStatus.Expired {
		log.WithField("polling_code", pollingCode).Debug("Polling code expired")
		return &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.APIErrorPollingCodeExpired,
		}
	}
	token, err := pollingcoderepo.PopTokenForPollingCode(nil, pollingCode)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if token == "" {
		return &model.Response{
			Status:   fiber.StatusPreconditionRequired,
			Response: model.APIErrorAuthorizationPending,
		}
	}
	st, err := supertoken.ParseJWT(token)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	log.Tracef("The JWT was parsed as '%+v'", st)
	pollingInfo := authcode.ParsePollingCode(pollingCode)
	res, err := st.ToTokenResponse(pollingInfo.ResponseType, networkData, token)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: res,
	}
}
