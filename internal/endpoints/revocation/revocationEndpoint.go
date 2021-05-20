package revocation

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/pkg/api/v0"
	sharedModel "github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/mytoken"
	mytokenPkg "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/token"
	"github.com/oidc-mytoken/server/shared/utils"
)

// HandleRevoke handles requests to the revocation endpoint
func HandleRevoke(ctx *fiber.Ctx) error {
	log.Debug("Handle revocation request")
	req := api.RevocationRequest{}
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	log.Trace("Parsed mytoken request")
	clearCookie := false
	if req.Token == "" {
		req.Token = ctx.Cookies("mytoken")
		clearCookie = true
	}
	errRes := revokeAnyToken(nil, req.Token, req.OIDCIssuer, req.Recursive)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	if clearCookie {
		return model.Response{
			Status: fiber.StatusNoContent,
			Cookies: []*fiber.Cookie{{
				Name:     "mytoken",
				Value:    "",
				Path:     "/api",
				Expires:  time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				Secure:   config.Get().Server.Secure,
				HTTPOnly: true,
				SameSite: "Strict",
			}},
		}.Send(ctx)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func revokeAnyToken(tx *sqlx.Tx, token, issuer string, recursive bool) (errRes *model.Response) {
	if utils.IsJWT(token) { // normal Mytoken
		return revokeMytoken(tx, token, issuer, recursive)
	} else if len(token) == config.Get().Features.Polling.Len { // Transfer Code
		return revokeTransferCode(tx, token, issuer)
	} else { // Short Token
		shortToken := transfercoderepo.ParseShortToken(token)
		var valid bool
		if err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
			jwt, v, err := shortToken.JWT(tx)
			valid = v
			if err != nil {
				return err
			}
			token = jwt
			return shortToken.Delete(tx)
		}); err != nil {
			return model.ErrorToInternalServerErrorResponse(err)
		}
		if !valid {
			return nil
		}
		return revokeMytoken(tx, token, issuer, recursive)
	}
}

func revokeMytoken(tx *sqlx.Tx, jwt, issuer string, recursive bool) (errRes *model.Response) {
	mt, err := mytokenPkg.ParseJWT(jwt)
	if err != nil {
		return nil
	}
	if issuer != "" && mt.OIDCIssuer != issuer {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: sharedModel.BadRequestError("token not for specified issuer"),
		}
	}
	return mytoken.RevokeMytoken(tx, mt.ID, token.Token(jwt), recursive, mt.OIDCIssuer)
}

func revokeTransferCode(tx *sqlx.Tx, token, issuer string) (errRes *model.Response) {
	transferCode := transfercoderepo.ParseTransferCode(token)
	err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		revokeMT, err := transferCode.GetRevokeJWT(tx)
		if err != nil {
			return err
		}
		if revokeMT {
			jwt, valid, err := transferCode.JWT(tx)
			if err != nil {
				return err
			}
			if valid { // if !valid the jwt field could not decrypted correctly, so we can skip that, but still delete the TransferCode
				errRes = revokeAnyToken(tx, jwt, issuer, true)
				if errRes != nil {
					return fmt.Errorf("placeholder")
				}
			}
		}
		return transferCode.Delete(tx)
	})
	if err != nil && errRes == nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return
}
