package mytoken

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/cryptstore"
	dbhelper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/refreshtokenrepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	provider2 "github.com/oidc-mytoken/server/internal/oidc/provider"
	"github.com/oidc-mytoken/server/internal/oidc/revoke"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

const errResPlaceholder = "error_res"

// HandleMytokenFromTransferCode handles requests to return the mytoken for a transfer code
func HandleMytokenFromTransferCode(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle mytoken from transfercode")
	req := response.NewExchangeTransferCodeRequest()
	if err := errors.WithStack(json.Unmarshal(ctx.Body(), &req)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	rlog.Trace("Parsed request")
	var errorRes *model.Response = nil
	var tokenStr string
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			status, err := transfercoderepo.CheckTransferCode(rlog, tx, req.TransferCode)
			if err != nil {
				return err
			}
			if !status.Found {
				errorRes = &model.Response{
					Status:   fiber.StatusUnauthorized,
					Response: api.ErrorBadTransferCode,
				}
				return errors.New(errResPlaceholder)
			}
			if status.Expired {
				errorRes = &model.Response{
					Status:   fiber.StatusUnauthorized,
					Response: api.ErrorTransferCodeExpired,
				}
				return errors.New(errResPlaceholder)
			}
			tokenStr, err = transfercoderepo.PopTokenForTransferCode(
				rlog, tx, req.TransferCode, *ctxutils.ClientMetaData(ctx),
			)
			return err
		},
	); err != nil {
		if errorRes != nil {
			return errorRes
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}

	token, err := universalmytoken.Parse(rlog, tokenStr)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToBadRequestErrorResponse(err)
	}
	mt, err := mytoken.ParseJWT(token.JWT)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{
		Status: fiber.StatusOK,
		Response: response.MytokenResponse{
			MytokenResponse: api.MytokenResponse{
				Mytoken:      token.OriginalToken,
				ExpiresIn:    mt.ExpiresIn(),
				Capabilities: mt.Capabilities,
				MOMID:        mt.ID.Hash(),
			},
			MytokenType:  token.OriginalTokenType,
			Restrictions: mt.Restrictions,
		},
	}

}

// RevokeMytoken revokes a Mytoken
func RevokeMytoken(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID, jwt string, recursive bool, issuer string,
) *model.Response {
	p := provider2.GetProvider(issuer)
	if p == nil {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}
	}
	err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			rtID, err := refreshtokenrepo.GetRTID(rlog, tx, id)
			if err != nil {
				_, err = db.ParseError(err) // sets err to nil if token was not found;
				// this is no error and we are done, since the token is already revoked
				return err
			}
			rt, _, err := cryptstore.GetRefreshToken(rlog, tx, id, jwt)
			if err != nil {
				return err
			}
			if err = dbhelper.RevokeMT(rlog, tx, id, recursive); err != nil {
				return err
			}
			count, err := refreshtokenrepo.CountRTOccurrences(rlog, tx, rtID)
			if err != nil {
				return err
			}
			if count > 0 {
				return nil
			}
			revoke.RefreshToken(rlog, p, rt)
			return cryptstore.DeleteCrypted(rlog, tx, rtID)
		},
	)
	if err == nil {
		return nil
	}
	rlog.Errorf("%s", errorfmt.Full(err))
	return model.ErrorToInternalServerErrorResponse(err)
}
