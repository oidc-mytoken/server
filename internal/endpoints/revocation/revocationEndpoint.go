package revocation

import (
	"fmt"
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
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/shorttokenrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	pkg2 "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytokenPkg "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/rotation"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleRevoke handles requests to the revocation endpoint
func HandleRevoke(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle revocation request")
	req := api.RevocationRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	rlog.WithField("parsed request", fmt.Sprintf("%+v", req)).WithField(
		"body", string(ctx.Body()),
	).Trace("Parsed revocation request")
	clearCookie := false
	if req.Token == "" {
		req.Token = ctx.Cookies("mytoken")
		if req.Token == "" {
			return model.BadRequestErrorResponse("no token given")
		}
		if req.MOMID == "" {
			clearCookie = true
		}
	}
	if req.MOMID != "" {
		var res *model.Response
		_ = db.Transact(
			rlog, func(tx *sqlx.Tx) error {
				metadata := ctxutils.ClientMetaData(ctx)
				token, err := universalmytoken.Parse(rlog, req.Token)
				if err != nil {
					res = model.ErrorToBadRequestErrorResponse(err)
					return err
				}
				authToken, err := mytokenPkg.ParseJWT(token.JWT)
				if err != nil {
					res = model.ErrorToBadRequestErrorResponse(err)
					return err
				}
				errRes := revokeByID(rlog, tx, req, authToken, metadata)
				if errRes != nil {
					res = errRes
					return errors.New("rollback")
				}
				tokenUpdate, err := rotation.RotateMytokenAfterOtherForResponse(
					rlog, tx, token.JWT, authToken, *metadata, token.OriginalTokenType,
				)
				if err != nil {
					res = model.ErrorToInternalServerErrorResponse(err)
					return err
				}
				if tokenUpdate != nil {
					res = &model.Response{
						Status: fiber.StatusOK,
						Response: pkg.OnlyTokenUpdateRes{
							TokenUpdate: tokenUpdate,
						},
						Cookies: []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)},
					}
				}
				return nil
			},
		)
		if res != nil {
			return res
		}
		return &model.Response{Status: fiber.StatusNoContent}
	}
	errRes := revokeAnyToken(rlog, nil, req.Token, req.OIDCIssuer, req.Recursive)
	if errRes != nil {
		return errRes
	}
	if clearCookie {
		return &model.Response{
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
		}
	}
	return &model.Response{Status: fiber.StatusNoContent}
}

func revokeByID(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req api.RevocationRequest,
	authToken *mytokenPkg.Mytoken,
	clientMetadata *api.ClientMetaData,
) (errRes *model.Response) {
	rollback := errors.New("rollback")
	_ = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			isParent, err := helper.MOMIDHasParent(rlog, nil, req.MOMID, authToken.ID)
			if err != nil {
				errRes = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if !isParent && !authToken.Capabilities.Has(api.CapabilityRevokeAnyToken) {
				errRes = &model.Response{
					Status: fiber.StatusForbidden,
					Response: api.Error{
						Error: api.ErrorStrInsufficientCapabilities,
						ErrorDescription: fmt.Sprintf(
							"The provided token is neither a parent of the token to be revoked"+
								" nor does it have the '%s' capability", api.CapabilityRevokeAnyToken.Name,
						),
					},
				}
				return rollback
			}
			same, err := helper.CheckMytokensAreForSameUser(rlog, nil, req.MOMID, authToken.ID)
			if err != nil {
				errRes = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if !same {
				errRes = &model.Response{
					Status: fiber.StatusForbidden,
					Response: api.Error{
						Error:            api.ErrorStrInvalidGrant,
						ErrorDescription: "The provided token cannot be used to revoke this mom_id",
					},
				}
				return rollback
			}
			if err = helper.RevokeMT(rlog, tx, req.MOMID, req.Recursive); err != nil {
				errRes = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if err = eventService.LogEvent(
				rlog, tx, pkg2.MTEvent{
					Event:          api.EventRevokedOtherToken,
					MTID:           authToken.ID,
					Comment:        fmt.Sprintf("mom_id: %s", req.MOMID),
					ClientMetaData: *clientMetadata,
				},
			); err != nil {
				errRes = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			return nil
		},
	)
	return
}

func revokeAnyToken(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, token, issuer string, recursive bool,
) (errRes *model.Response) {
	if jwtutils.IsJWT(token) { // normal Mytoken
		return revokeMytoken(rlog, tx, token, issuer, recursive)
	} else if len(token) < api.MinShortTokenLen { // Transfer Code
		return revokeTransferCode(rlog, tx, token, issuer)
	} else { // Short Token
		shortToken := shorttokenrepo.ParseShortToken(token)
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
		return model.BadRequestErrorResponse("token not for specified issuer")
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
