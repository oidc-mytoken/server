package polling

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandlePollingCode handles a request on the polling endpoint
func HandlePollingCode(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	req := response.NewPollingCodeRequest()
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	clientMetaData := ctxutils.ClientMetaData(ctx)
	mt, token, pollingCodeStatus, errRes := CheckPollingCodeReq(rlog, req, *clientMetaData, false)
	if errRes != nil {
		return errRes
	}
	maxTokenLen := 0
	if pollingCodeStatus.MaxTokenLen != nil {
		maxTokenLen = *pollingCodeStatus.MaxTokenLen
	}
	res, err := mt.ToTokenResponse(rlog, pollingCodeStatus.ResponseType, maxTokenLen, *clientMetaData, token)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: res,
	}
}

// CheckPollingCodeReq checks a pkg.PollingCodeRequest and returns the linked mytoken if valid
func CheckPollingCodeReq(
	rlog log.Ext1FieldLogger, req response.PollingCodeRequest, networkData api.ClientMetaData, forSSH bool,
) (mt *mytoken.Mytoken, token string, pollingCodeStatus transfercoderepo.TransferCodeStatus, errRes *model.Response) {
	pollingCode := req.PollingCode
	rlog.WithField("polling_code", pollingCode).WithField("for_ssh", forSSH).Debug("Handle polling code")
	var err error
	pollingCodeStatus, err = transfercoderepo.CheckTransferCode(rlog, nil, pollingCode)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		errRes = model.ErrorToInternalServerErrorResponse(err)
		return
	}
	if !pollingCodeStatus.Found || (pollingCodeStatus.SSHKeyFingerprint.Valid != forSSH) {
		rlog.WithField("polling_code", pollingCode).Debug("Polling code not known")
		errRes = &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: api.ErrorBadTransferCode,
		}
		return
	}
	if pollingCodeStatus.ConsentDeclined {
		rlog.WithField("polling_code", pollingCode).Debug("Consent declined")
		errRes = &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: api.ErrorConsentDeclined,
		}
		return
	}
	if pollingCodeStatus.Expired {
		rlog.WithField("polling_code", pollingCode).Debug("Polling code expired")
		errRes = &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: api.ErrorTransferCodeExpired,
		}
		return
	}
	token, err = transfercoderepo.PopTokenForTransferCode(rlog, nil, pollingCode, networkData)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
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
	mt, err = mytoken.ParseJWTWithoutClaimsValidation(token)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		errRes = model.ErrorToInternalServerErrorResponse(err)
		return
	}
	return
}
