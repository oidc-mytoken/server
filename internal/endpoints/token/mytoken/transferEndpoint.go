package mytoken

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxUtils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	pkgModel "github.com/oidc-mytoken/server/shared/model"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
)

// HandleCreateTransferCodeForExistingMytoken handles request to create a transfer code for an existing mytoken
func HandleCreateTransferCodeForExistingMytoken(ctx *fiber.Ctx) error {
	var req pkg.CreateTransferCodeRequest
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.Response{
			Status:   fiber.StatusBadRequest,
			Response: pkgModel.BadRequestError(errorfmt.Error(err)),
		}.Send(ctx)
	}
	mt, errRes := auth.RequireValidMytoken(nil, &req.Mytoken, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}

	transferCode, expiresIn, err := mytoken.CreateTransferCode(mt.ID, req.Mytoken.JWT, false, req.Mytoken.OriginalTokenType, *ctxUtils.ClientMetaData(ctx))
	if err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	res := &model.Response{
		Status: fiber.StatusOK,
		Response: pkg.TransferCodeResponse{
			MytokenType: pkgModel.ResponseTypeTransferCode,
			TransferCodeResponse: api.TransferCodeResponse{
				TransferCode: transferCode,
				ExpiresIn:    expiresIn,
			},
		},
	}
	return res.Send(ctx)
}
