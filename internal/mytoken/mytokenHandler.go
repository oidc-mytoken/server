package mytoken

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/cryptstore"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/encryptionkeyrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo"
	dbhelper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/refreshtokenrepo"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	"github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/rotation"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	provider2 "github.com/oidc-mytoken/server/internal/oidc/provider"
	"github.com/oidc-mytoken/server/internal/oidc/revoke"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
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

// HandleMytokenFromMytokenReqChecks handles the necessary req checks for a pkg.MytokenFromMytokenRequest
func HandleMytokenFromMytokenReqChecks(
	rlog log.Ext1FieldLogger, req *response.MytokenFromMytokenRequest, clientData *api.ClientMetaData,
	ctx *fiber.Ctx,
) (*restrictions.Restriction, *mytoken.Mytoken, *model.Response) {
	req.Restrictions.ReplaceThisIP(clientData.IP)
	req.Restrictions.ClearUnsupportedKeys()
	rlog.Trace("Parsed mytoken request")

	// GrantType already checked

	mt, errRes := auth.RequireValidMytoken(rlog, nil, &req.Mytoken, ctx)
	if errRes != nil {
		return nil, nil, errRes
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt, clientData, api.CapabilityCreateMT,
	)
	if errRes != nil {
		return nil, nil, errRes
	}
	if _, errRes = auth.RequireMatchingIssuer(rlog, mt.OIDCIssuer, &req.GeneralMytokenRequest.Issuer); errRes != nil {
		return nil, nil, errRes
	}
	return usedRestriction, mt, nil
}

// HandleMytokenFromMytoken handles requests to create a Mytoken from an existing Mytoken
func HandleMytokenFromMytoken(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle mytoken from mytoken")
	req := response.NewMytokenRequest()
	if err := errors.WithStack(json.Unmarshal(ctx.Body(), &req)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	usedRestriction, mt, errRes := HandleMytokenFromMytokenReqChecks(rlog, req, ctxutils.ClientMetaData(ctx), ctx)
	if errRes != nil {
		return errRes
	}
	return HandleMytokenFromMytokenReq(rlog, mt, req, ctxutils.ClientMetaData(ctx), usedRestriction)
}

// HandleMytokenFromMytokenReq handles a mytoken request (from an existing mytoken)
func HandleMytokenFromMytokenReq(
	rlog log.Ext1FieldLogger, parent *mytoken.Mytoken, req *response.MytokenFromMytokenRequest,
	networkData *api.ClientMetaData,
	usedRestriction *restrictions.Restriction,
) *model.Response {
	ste, errorResponse := createMytokenEntry(rlog, parent, req, *networkData)
	if errorResponse != nil {
		return errorResponse
	}
	var tokenUpdate *response.MytokenResponse
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) (err error) {
			if usedRestriction != nil {
				if err = usedRestriction.UsedOther(rlog, tx, parent.ID); err != nil {
					return
				}
			}
			tokenUpdate, err = rotation.RotateMytokenAfterOtherForResponse(
				rlog, tx, req.Mytoken.JWT, parent, *networkData, req.Mytoken.OriginalTokenType,
			)
			if err != nil {
				return
			}
			if err = ste.Store(rlog, tx, "Used grant_type mytoken"); err != nil {
				return
			}
			if err = notificationsrepo.ExpandNotificationsToChildrenIfApplicable(
				rlog, tx, parent.ID, ste.ID,
			); err != nil {
				return err
			}
			if err = notificationsrepo.ScheduleExpirationNotificationsIfNeeded(
				rlog, tx, ste.ID, ste.Token.ExpiresAt, ste.Token.IssuedAt,
			); err != nil {
				return err
			}
			return eventService.LogEvents(
				rlog, tx, []pkg.MTEvent{
					{
						Event:          api.EventInheritedRT,
						Comment:        "Got RT from parent",
						MTID:           ste.ID,
						ClientMetaData: *networkData,
					},
					{
						Event:          api.EventSubtokenCreated,
						Comment:        strings.TrimSpace(fmt.Sprintf("Created MT %s", req.GeneralMytokenRequest.Name)),
						MTID:           parent.ID,
						ClientMetaData: *networkData,
					},
				},
			)
		},
	); err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}

	res, err := ste.Token.ToTokenResponse(
		rlog, req.ResponseType, req.GeneralMytokenRequest.MaxTokenLen, *networkData, "",
	)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	var cake []*fiber.Cookie
	if tokenUpdate != nil {
		res.TokenUpdate = tokenUpdate
		cake = []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)}
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: res,
		Cookies:  cake,
	}
}

func createMytokenEntry(
	rlog log.Ext1FieldLogger, parent *mytoken.Mytoken, req *response.MytokenFromMytokenRequest,
	networkData api.ClientMetaData,
) (*mytokenrepo.MytokenEntry, *model.Response) {
	rtID, dbErr := refreshtokenrepo.GetRTID(rlog, nil, parent.ID)
	rtFound, err := db.ParseError(dbErr)
	if err != nil {
		rlog.WithError(dbErr).Error()
		return nil, model.ErrorToInternalServerErrorResponse(dbErr)
	}
	if !rtFound {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.InvalidTokenError(""),
		}
	}
	if changed := req.Restrictions.EnforceMaxLifetime(parent.OIDCIssuer); changed && req.FailOnRestrictionsNotTighter {
		return nil, model.BadRequestErrorResponse("requested restrictions do not respect maximum mytoken lifetime")
	}
	r, ok := restrictions.Tighten(rlog, parent.Restrictions, req.Restrictions.Restrictions)
	if !ok && req.FailOnRestrictionsNotTighter {
		return nil, model.BadRequestErrorResponse("requested restrictions are not subset of original restrictions")
	}
	c := api.TightenCapabilities(parent.Capabilities, req.Capabilities.Capabilities)
	if len(c) == 0 {
		return nil, model.BadRequestErrorResponse("mytoken to be issued cannot have any of the requested capabilities")
	}
	var rot *api.Rotation
	if req.Rotation != nil {
		rot = &req.Rotation.Rotation
	}
	mt, err := mytoken.NewMytoken(
		parent.OIDCSubject, parent.OIDCIssuer, req.GeneralMytokenRequest.Name, r, c, rot,
		parent.AuthTime,
	)
	if err != nil {
		return nil, model.ErrorToInternalServerErrorResponse(err)
	}
	mte := mytokenrepo.NewMytokenEntry(mt, req.GeneralMytokenRequest.Name, networkData)
	encryptionKey, _, err := encryptionkeyrepo.GetEncryptionKey(rlog, nil, parent.ID, req.Mytoken.JWT)
	if err != nil {
		rlog.WithError(err).Error()
		return mte, model.ErrorToInternalServerErrorResponse(err)
	}
	if err = mte.SetRefreshToken(rtID, encryptionKey); err != nil {
		rlog.WithError(err).Error()
		return mte, model.ErrorToInternalServerErrorResponse(err)
	}
	mte.ParentID = parent.ID
	return mte, nil
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
