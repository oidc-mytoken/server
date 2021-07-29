package mytoken

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	dbhelper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/ctxUtils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	pkgModel "github.com/oidc-mytoken/server/shared/model"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/utils"
)

// HandleCreateTransferCodeForExistingMytoken handles request to create a transfer code for an existing mytoken
func HandleCreateTransferCodeForExistingMytoken(ctx *fiber.Ctx) error {
	token := ctxUtils.GetAuthHeaderToken(ctx)
	if token == "" {
		var req api.CreateTransferCodeRequest
		if err := json.Unmarshal(ctx.Body(), &req); err != nil {
			return model.Response{
				Status:   fiber.StatusBadRequest,
				Response: pkgModel.BadRequestError(errorfmt.Error(err)),
			}.Send(ctx)
		}
		token = req.Mytoken
		if token == "" {
			return model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: pkgModel.BadRequestError("required parameter 'mytoken' missing"),
			}.Send(ctx)
		}
	}

	var jwt string
	var tokenType pkgModel.ResponseType
	if utils.IsJWT(token) {
		jwt = token
		tokenType = pkgModel.ResponseTypeToken
	} else {
		tokenType = pkgModel.ResponseTypeShortToken
		shortToken := transfercoderepo.ParseShortToken(token)
		tmp, valid, err := shortToken.JWT(nil)
		jwt = tmp
		if err != nil {
			return model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: pkgModel.InvalidTokenError(errorfmt.Error(err)),
			}.Send(ctx)
		}
		if !valid {
			return model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: pkgModel.InvalidTokenError(""),
			}.Send(ctx)
		}
	}
	mToken, err := mytoken.ParseJWT(jwt)
	if err != nil {
		return model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: pkgModel.InvalidTokenError(errorfmt.Error(err)),
		}.Send(ctx)
	}

	revoked, dbErr := dbhelper.CheckTokenRevoked(nil, mToken.ID, mToken.SeqNo, mToken.Rotation)
	if dbErr != nil {
		log.Errorf("%s", errorfmt.Full(dbErr))
		return model.ErrorToInternalServerErrorResponse(dbErr).Send(ctx)
	}
	if revoked {
		return model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: pkgModel.InvalidTokenError(""),
		}.Send(ctx)
	}

	transferCode, expiresIn, err := mytoken.CreateTransferCode(mToken.ID, token, false, tokenType, *ctxUtils.ClientMetaData(ctx))
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
