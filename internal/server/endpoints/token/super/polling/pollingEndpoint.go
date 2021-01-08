package polling

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/server/db/dbrepo/supertokenrepo/transfercoderepo"
	response "github.com/zachmann/mytoken/internal/server/endpoints/token/super/pkg"
	"github.com/zachmann/mytoken/internal/server/model"
	supertoken "github.com/zachmann/mytoken/internal/server/supertoken/pkg"
	"github.com/zachmann/mytoken/internal/server/utils/ctxUtils"
	pkgModel "github.com/zachmann/mytoken/pkg/model"
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
	pollingCodeStatus, err := transfercoderepo.CheckTransferCode(nil, pollingCode)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if !pollingCodeStatus.Found {
		log.WithField("polling_code", pollingCode).Debug("Polling code not known")
		return &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: pkgModel.APIErrorBadTransferCode,
		}
	}
	if pollingCodeStatus.Expired {
		log.WithField("polling_code", pollingCode).Debug("Polling code expired")
		return &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: pkgModel.APIErrorTransferCodeExpired,
		}
	}
	token, err := transfercoderepo.PopTokenForTransferCode(nil, pollingCode)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if token == "" {
		return &model.Response{
			Status:   fiber.StatusPreconditionRequired,
			Response: pkgModel.APIErrorAuthorizationPending,
		}
	}
	st, err := supertoken.ParseJWT(token)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	log.Tracef("The JWT was parsed as '%+v'", st)
	res, err := st.ToTokenResponse(pollingCodeStatus.ResponseType, networkData, token)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: res,
	}
}
