package mytoken

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/encryptionkeyrepo"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/shared/mytoken/rotation"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo"
	dbhelper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/refreshtokenrepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/revoke"
	"github.com/oidc-mytoken/server/internal/server/httpStatus"
	"github.com/oidc-mytoken/server/internal/utils/ctxUtils"
	pkgModel "github.com/oidc-mytoken/server/shared/model"
	eventService "github.com/oidc-mytoken/server/shared/mytoken/event"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
	"github.com/oidc-mytoken/server/shared/mytoken/universalmytoken"
)

const errResPlaceholder = "error_res"

// HandleMytokenFromTransferCode handles requests to return the mytoken for a transfer code
func HandleMytokenFromTransferCode(ctx *fiber.Ctx) *model.Response {
	log.Debug("Handle mytoken from transfercode")
	req := response.NewExchangeTransferCodeRequest()
	if err := errors.WithStack(json.Unmarshal(ctx.Body(), &req)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	log.Trace("Parsed request")
	var errorRes *model.Response = nil
	var tokenStr string
	if err := db.Transact(func(tx *sqlx.Tx) error {
		status, err := transfercoderepo.CheckTransferCode(tx, req.TransferCode)
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
		tokenStr, err = transfercoderepo.PopTokenForTransferCode(tx, req.TransferCode, *ctxUtils.ClientMetaData(ctx))
		return err
	}); err != nil {
		if errorRes != nil {
			return errorRes
		}
		log.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}

	token, err := universalmytoken.Parse(tokenStr)
	if err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	mt, err := mytoken.ParseJWT(token.JWT)
	if err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{
		Status: fiber.StatusOK,
		Response: response.MytokenResponse{
			MytokenResponse: api.MytokenResponse{
				Mytoken:              token.OriginalToken,
				ExpiresIn:            mt.ExpiresIn(),
				Capabilities:         mt.Capabilities,
				SubtokenCapabilities: mt.SubtokenCapabilities,
			},
			MytokenType:  token.OriginalTokenType,
			Restrictions: mt.Restrictions,
		},
	}

}

// HandleMytokenFromMytoken handles requests to create a Mytoken from an existing Mytoken
func HandleMytokenFromMytoken(ctx *fiber.Ctx) *model.Response {
	log.Debug("Handle mytoken from mytoken")
	req := response.NewMytokenRequest()
	if err := errors.WithStack(json.Unmarshal(ctx.Body(), &req)); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	if req.Capabilities != nil && len(req.Capabilities) == 0 {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.Error{Error: api.ErrorStrInvalidRequest, ErrorDescription: "capabilities cannot be empty"},
		}
	}
	req.Restrictions.ReplaceThisIp(ctx.IP())
	req.Restrictions.ClearUnsupportedKeys()
	log.Trace("Parsed mytoken request")

	// GrantType already checked

	if req.Mytoken.JWT == "" {
		var err error
		req.Mytoken, err = universalmytoken.Parse(ctx.Cookies("mytoken"))
		if err != nil {
			return &model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: pkgModel.InvalidTokenError(errorfmt.Error(err)),
			}
		}
	}

	mt, err := mytoken.ParseJWT(req.Mytoken.JWT)
	if err != nil {
		return &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: pkgModel.InvalidTokenError(errorfmt.Error(err)),
		}
	}
	log.Trace("Parsed mytoken")

	revoked, dbErr := dbhelper.CheckTokenRevoked(nil, mt.ID, mt.SeqNo, mt.Rotation)
	if dbErr != nil {
		return model.ErrorToInternalServerErrorResponse(dbErr)
	}
	if revoked {
		return &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: pkgModel.InvalidTokenError(""),
		}
	}
	log.Trace("Checked token not revoked")

	if ok := mt.VerifyCapabilities(api.CapabilityCreateMT); !ok {
		return &model.Response{
			Status:   fiber.StatusForbidden,
			Response: api.ErrorInsufficientCapabilities,
		}
	}
	log.Trace("Checked mytoken capabilities")
	if ok := mt.Restrictions.VerifyForOther(nil, ctx.IP(), mt.ID); !ok {
		return &model.Response{
			Status:   fiber.StatusForbidden,
			Response: api.ErrorUsageRestricted,
		}
	}
	log.Trace("Checked mytoken restrictions")

	if req.Issuer == "" {
		req.Issuer = mt.OIDCIssuer
	} else {
		if req.Issuer != mt.OIDCIssuer {
			return &model.Response{
				Status:   fiber.StatusBadRequest,
				Response: pkgModel.BadRequestError("token not for specified issuer"),
			}
		}
		log.Trace("Checked issuer")
	}
	return handleMytokenFromMytoken(mt, req, ctxUtils.ClientMetaData(ctx))
}

func handleMytokenFromMytoken(parent *mytoken.Mytoken, req *response.MytokenFromMytokenRequest, networkData *api.ClientMetaData) *model.Response {
	ste, errorResponse := createMytokenEntry(parent, req, *networkData)
	if errorResponse != nil {
		return errorResponse
	}
	var tokenUpdate *response.MytokenResponse
	if err := db.Transact(func(tx *sqlx.Tx) (err error) {
		if len(parent.Restrictions) > 0 {
			if err = parent.Restrictions.GetValidForOther(tx, networkData.IP, parent.ID)[0].UsedOther(tx, parent.ID); err != nil {
				return
			}
		}
		tokenUpdate, err = rotation.RotateMytokenAfterOtherForResponse(
			tx, req.Mytoken.JWT, parent, *networkData, req.Mytoken.OriginalTokenType)
		if err != nil {
			return
		}
		if err = ste.Store(tx, "Used grant_type mytoken"); err != nil {
			return
		}
		return eventService.LogEvents(tx, []eventService.MTEvent{
			{Event: event.FromNumber(event.MTEventInheritedRT, "Got RT from parent"), MTID: ste.ID},
			{Event: event.FromNumber(event.MTEventMTCreated, strings.TrimSpace(fmt.Sprintf("Created MT %s", req.Name))), MTID: parent.ID},
		}, *networkData)
	}); err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}

	res, err := ste.Token.ToTokenResponse(req.ResponseType, req.MaxTokenLen, *networkData, "")
	if err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	var cake []*fiber.Cookie
	if tokenUpdate != nil {
		res.TokenUpdate = tokenUpdate
		cookie := cookies.MytokenCookie(tokenUpdate.Mytoken)
		cake = []*fiber.Cookie{&cookie}
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: res,
		Cookies:  cake,
	}
}

func createMytokenEntry(parent *mytoken.Mytoken, req *response.MytokenFromMytokenRequest, networkData api.ClientMetaData) (*mytokenrepo.MytokenEntry, *model.Response) {
	rtID, dbErr := refreshtokenrepo.GetRTID(nil, parent.ID)
	rtFound, err := dbhelper.ParseError(dbErr)
	if err != nil {
		log.WithError(dbErr).Error()
		return nil, model.ErrorToInternalServerErrorResponse(dbErr)
	}
	if !rtFound {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: pkgModel.InvalidTokenError(""),
		}
	}
	rootID, rootFound, dbErr := dbhelper.GetMTRootID(parent.ID)
	if dbErr != nil {
		log.WithError(dbErr).Error()
		return nil, model.ErrorToInternalServerErrorResponse(dbErr)
	}
	if !rootFound {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: pkgModel.InvalidTokenError(""),
		}
	}
	if !rootID.HashValid() {
		rootID = parent.ID
	}
	if changed := req.Restrictions.EnforceMaxLifetime(parent.OIDCIssuer); changed && req.FailOnRestrictionsNotTighter {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: pkgModel.BadRequestError("requested restrictions do not respect maximum mytoken lifetime"),
		}
	}
	r, ok := restrictions.Tighten(parent.Restrictions, req.Restrictions)
	if !ok && req.FailOnRestrictionsNotTighter {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: pkgModel.BadRequestError("requested restrictions are not subset of original restrictions"),
		}
	}
	capsFromParent := parent.SubtokenCapabilities
	if capsFromParent == nil {
		capsFromParent = parent.Capabilities
	}
	c := api.TightenCapabilities(capsFromParent, req.Capabilities)
	if len(c) == 0 {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: pkgModel.BadRequestError("mytoken to be issued cannot have any of the requested capabilities"),
		}
	}
	var sc api.Capabilities = nil
	if c.Has(api.CapabilityCreateMT) {
		sc = api.TightenCapabilities(capsFromParent, req.SubtokenCapabilities)
	}
	ste := mytokenrepo.NewMytokenEntry(
		mytoken.NewMytoken(parent.OIDCSubject, parent.OIDCIssuer, r, c, sc, req.Rotation),
		req.Name, networkData)
	encryptionKey, _, err := encryptionkeyrepo.GetEncryptionKey(nil, parent.ID, req.Mytoken.JWT)
	if err != nil {
		log.WithError(err).Error()
		return ste, model.ErrorToInternalServerErrorResponse(err)
	}
	if err = ste.SetRefreshToken(rtID, encryptionKey); err != nil {
		log.WithError(err).Error()
		return ste, model.ErrorToInternalServerErrorResponse(err)
	}
	ste.ParentID = parent.ID
	ste.RootID = rootID
	return ste, nil
}

// RevokeMytoken revokes a Mytoken
func RevokeMytoken(tx *sqlx.Tx, id mtid.MTID, jwt string, recursive bool, issuer string) *model.Response {
	provider, ok := config.Get().ProviderByIssuer[issuer]
	if !ok {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}
	}
	err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		rtID, err := refreshtokenrepo.GetRTID(tx, id)
		if err != nil {
			_, err = dbhelper.ParseError(err) // sets err to nil if token was not found;
			// this is no error and we are done, since the token is already revoked
			return err
		}
		rt, _, err := refreshtokenrepo.GetRefreshToken(tx, id, jwt)
		if err != nil {
			return err
		}
		if err = dbhelper.RevokeMT(tx, id, recursive); err != nil {
			return err
		}
		count, err := refreshtokenrepo.CountRTOccurrences(tx, rtID)
		if err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		if e := revoke.RefreshToken(provider, rt); e != nil {
			apiError := e.Response.(api.Error)
			return fmt.Errorf("%s: %s", apiError.Error, apiError.ErrorDescription)
		}
		return refreshtokenrepo.DeleteRefreshToken(tx, rtID)
	})
	if err == nil {
		return nil
	}
	if strings.HasPrefix(errorfmt.Error(err), "oidc_error") {
		return &model.Response{
			Status:   httpStatus.StatusOIDPError,
			Response: pkgModel.OIDCError(errorfmt.Error(err), ""),
		}
	}
	log.Errorf("%s", errorfmt.Full(err))
	return model.ErrorToInternalServerErrorResponse(err)
}
