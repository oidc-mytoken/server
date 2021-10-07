package polling

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/ctxUtils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
)

// HandlePollingCode handles a request on the polling endpoint
func HandlePollingCode(ctx *fiber.Ctx) error {
	req := response.NewPollingCodeRequest()
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	clientMetaData := ctxUtils.ClientMetaData(ctx)
	mt, token, pollingCodeStatus, errRes := CheckPollingCodeReq(req, *clientMetaData, false)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	maxTokenLen := 0
	if pollingCodeStatus.MaxTokenLen != nil {
		maxTokenLen = *pollingCodeStatus.MaxTokenLen
	}
	res, err := mt.ToTokenResponse(pollingCodeStatus.ResponseType, maxTokenLen, *clientMetaData, token)
	if err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return model.Response{
		Status:   fiber.StatusOK,
		Response: res,
	}.Send(ctx)
}

func CheckPollingCodeReq(
	req response.PollingCodeRequest,
	networkData api.ClientMetaData,
	forSSH bool,
) (mt *mytoken.Mytoken, token string, pollingCodeStatus transfercoderepo.TransferCodeStatus, errRes *model.Response) {
	pollingCode := req.PollingCode
	log.WithField("polling_code", pollingCode).WithField("for_ssh", forSSH).Debug("Handle polling code")
	var err error
	pollingCodeStatus, err = transfercoderepo.CheckTransferCode(nil, pollingCode)
	if err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		errRes = model.ErrorToInternalServerErrorResponse(err)
		return
	}
	if !pollingCodeStatus.Found || (pollingCodeStatus.SSHKeyHash.Valid != forSSH) {
		log.WithField("polling_code", pollingCode).Debug("Polling code not known")
		errRes = &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: api.ErrorBadTransferCode,
		}
		return
	}
	if pollingCodeStatus.ConsentDeclined {
		log.WithField("polling_code", pollingCode).Debug("Consent declined")
		errRes = &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: api.ErrorConsentDeclined,
		}
		return
	}
	if pollingCodeStatus.Expired {
		log.WithField("polling_code", pollingCode).Debug("Polling code expired")
		errRes = &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: api.ErrorTransferCodeExpired,
		}
		return
	}
	token, err = transfercoderepo.PopTokenForTransferCode(nil, pollingCode, networkData)
	if err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		errRes = model.ErrorToInternalServerErrorResponse(err)
		return
	}
	if token == "" {
		errRes = &model.Response{
			Status:   fiber.StatusPreconditionRequired,
			Response: api.ErrorAuthorizationPending,
		}
		return
	}
	mt, err = mytoken.ParseJWT(token)
	if err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		errRes = model.ErrorToInternalServerErrorResponse(err)
		return
	}
	return
}
