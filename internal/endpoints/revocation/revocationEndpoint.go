package revocation

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/utils/utils/jwtutils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	helper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken"
	mytokenPkg "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleRevoke handles requests to the revocation endpoint
func HandleRevoke(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle revocation request")
	req := api.RevocationRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	rlog.Trace("Parsed mytoken request")
	clearCookie := false
	if req.Token == "" {
		req.Token = ctx.Cookies("mytoken")
		if req.RevocationID == "" {
			clearCookie = true
		}
	}
	if req.RevocationID != "" {
		errRes := revokeByID(rlog, req)
		if errRes != nil {
			return errRes.Send(ctx)
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	}
	errRes := revokeAnyToken(rlog, nil, req.Token, req.OIDCIssuer, req.Recursive)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	if clearCookie {
		return model.Response{
			Status: fiber.StatusNoContent,
			Cookies: []*fiber.Cookie{
				{
					Name:     "mytoken",
					Value:    "",
					Path:     "/api",
					Expires:  time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
					Secure:   config.Get().Server.Secure,
					HTTPOnly: true,
					SameSite: "Strict",
				},
			},
		}.Send(ctx)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func revokeByID(rlog log.Ext1FieldLogger, req api.RevocationRequest) (errRes *model.Response) {
	token, err := universalmytoken.Parse(rlog, req.Token)
	if err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	authToken, err := mytokenPkg.ParseJWT(token.JWT)
	if err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	isParent, err := helper.RevocationIDHasParent(rlog, nil, req.RevocationID, authToken.ID)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if !isParent && !authToken.Capabilities.Has(api.CapabilityRevokeAnyToken) {
		return &model.Response{
			Status: fiber.StatusForbidden,
			Response: api.Error{
				Error: api.ErrorStrInsufficientCapabilities,
				ErrorDescription: "The provided token is neither a parent of the the token to be revoked nor does it" +
					" have the 'revoke_any_token' capability",
			},
		}
	}
	same, err := helper.CheckMytokensAreForSameUser(rlog, nil, req.RevocationID, authToken.ID)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if !same {
		return &model.Response{
			Status: fiber.StatusForbidden,
			Response: api.Error{
				Error:            api.ErrorStrInvalidGrant,
				ErrorDescription: "The provided token cannot be used to revoke this revocation_id",
			},
		}
	}
	if err = helper.RevokeMT(rlog, nil, req.RevocationID, req.Recursive); err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return
}

func revokeAnyToken(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, token, issuer string, recursive bool,
) (errRes *model.Response) {
	if jwtutils.IsJWT(token) { // normal Mytoken
		return revokeMytoken(rlog, tx, token, issuer, recursive)
	} else if len(token) == config.Get().Features.Polling.Len { // Transfer Code
		return revokeTransferCode(rlog, tx, token, issuer)
	} else { // Short Token
		shortToken := transfercoderepo.ParseShortToken(token)
		var valid bool
		if err := db.RunWithinTransaction(
			rlog, tx, func(tx *sqlx.Tx) error {
				jwt, v, err := shortToken.JWT(rlog, tx)
				valid = v
				if err != nil {
					return err
				}
				token = jwt
				return shortToken.Delete(rlog, tx)
			},
		); err != nil {
			rlog.Errorf("%s", errorfmt.Full(err))
			return model.ErrorToInternalServerErrorResponse(err)
		}
		if !valid {
			return nil
		}
		return revokeMytoken(rlog, tx, token, issuer, recursive)
	}
}

func revokeMytoken(rlog log.Ext1FieldLogger, tx *sqlx.Tx, jwt, issuer string, recursive bool) (errRes *model.Response) {
	mt, err := mytokenPkg.ParseJWT(jwt)
	if err != nil {
		return nil
	}
	if issuer != "" && mt.OIDCIssuer != issuer {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError("token not for specified issuer"),
		}
	}
	return mytoken.RevokeMytoken(rlog, tx, mt.ID, jwt, recursive, mt.OIDCIssuer)
}

func revokeTransferCode(rlog log.Ext1FieldLogger, tx *sqlx.Tx, token, issuer string) (errRes *model.Response) {
	transferCode := transfercoderepo.ParseTransferCode(token)
	err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			revokeMT, err := transferCode.GetRevokeJWT(rlog, tx)
			if err != nil {
				return err
			}
			if revokeMT {
				jwt, valid, err := transferCode.JWT(rlog, tx)
				if err != nil {
					return err
				}
				if valid { // if !valid the jwt field could not be decrypted correctly, so we can skip that,
					// but still delete the TransferCode
					errRes = revokeAnyToken(rlog, tx, jwt, issuer, true)
					if errRes != nil {
						return errors.New("placeholder")
					}
				}
			}
			return transferCode.Delete(rlog, tx)
		},
	)
	if err != nil && errRes == nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return
}
