package revocation

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/db/dbrepo/supertokenrepo/transfercoderepo"
	request "github.com/zachmann/mytoken/internal/endpoints/revocation/pkg"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/supertoken"
	supertokenPkg "github.com/zachmann/mytoken/internal/supertoken/pkg"
	"github.com/zachmann/mytoken/internal/supertoken/token"
	pkgModel "github.com/zachmann/mytoken/pkg/model"
	"github.com/zachmann/mytoken/shared/utils"
)

// HandleRevoke handles requests to the revocation endpoint
func HandleRevoke(ctx *fiber.Ctx) error {
	log.Debug("Handle revocation request")
	req := request.RevocationRequest{}
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	log.Trace("Parsed super token request")

	errRes := revokeAnyToken(nil, req.Token, req.OIDCIssuer, req.Recursive)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func revokeAnyToken(tx *sqlx.Tx, token, issuer string, recursive bool) (errRes *model.Response) {
	if utils.IsJWT(token) { // normal SuperToken
		return revokeSuperToken(tx, token, issuer, recursive)
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
			if err = shortToken.Delete(tx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return model.ErrorToInternalServerErrorResponse(err)
		}
		if !valid {
			return nil
		}
		return revokeSuperToken(tx, token, issuer, recursive)
	}
}

func revokeSuperToken(tx *sqlx.Tx, jwt, issuer string, recursive bool) (errRes *model.Response) {
	st, err := supertokenPkg.ParseJWT(jwt)
	if err != nil {
		return nil
	}
	if len(issuer) > 0 && st.OIDCIssuer != issuer {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: pkgModel.BadRequestError("token not for specified issuer"),
		}
	}
	return supertoken.RevokeSuperToken(tx, st.ID, token.Token(jwt), recursive, st.OIDCIssuer)
}

func revokeTransferCode(tx *sqlx.Tx, token, issuer string) (errRes *model.Response) {
	transferCode := transfercoderepo.ParseTransferCode(token)
	err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		revokeST, err := transferCode.GetRevokeJWT(tx)
		if err != nil {
			return err
		}
		if revokeST {
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
