package mytoken

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

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
	pkgModel "github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/mytoken/capabilities"
	eventService "github.com/oidc-mytoken/server/shared/mytoken/event"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
	"github.com/oidc-mytoken/server/shared/mytoken/token"
	"github.com/oidc-mytoken/server/shared/utils"
)

// HandleMytokenFromTransferCode handles requests to return the mytoken for a transfer code
func HandleMytokenFromTransferCode(ctx *fiber.Ctx) *model.Response {
	log.Debug("Handle mytoken from transfercode")
	req := response.ExchangeTransferCodeRequest{}
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
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
				Response: pkgModel.APIErrorBadTransferCode,
			}
			return fmt.Errorf("error_res")
		}
		if status.Expired {
			errorRes = &model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: pkgModel.APIErrorTransferCodeExpired,
			}
			return fmt.Errorf("error_res")
		}
		tokenStr, err = transfercoderepo.PopTokenForTransferCode(tx, req.TransferCode, *ctxUtils.ClientMetaData(ctx))
		return err
	}); err != nil {
		if errorRes != nil {
			return errorRes
		}
		return model.ErrorToInternalServerErrorResponse(err)
	}

	tokenType := pkgModel.ResponseTypeToken
	if !utils.IsJWT(tokenStr) {
		tokenType = pkgModel.ResponseTypeShortToken
	}
	jwt, err := token.GetLongMytoken(tokenStr)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	mt, err := mytoken.ParseJWT(string(jwt))
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{
		Status: fiber.StatusOK,
		Response: response.MytokenResponse{
			Mytoken:              tokenStr,
			MytokenType:          tokenType,
			ExpiresIn:            mt.ExpiresIn(),
			Restrictions:         mt.Restrictions,
			Capabilities:         mt.Capabilities,
			SubtokenCapabilities: mt.SubtokenCapabilities,
		},
	}

}

// HandleMytokenFromMytoken handles requests to create a Mytoken from an existing Mytoken
func HandleMytokenFromMytoken(ctx *fiber.Ctx) *model.Response {
	log.Debug("Handle mytoken from mytoken")
	req := response.NewMytokenRequest()
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	log.Trace("Parsed mytoken request")

	// GrantType already checked

	if len(req.Mytoken) == 0 {
		var err error
		req.Mytoken, err = token.GetLongMytoken(ctx.Cookies("mytoken"))
		if err != nil {
			return &model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: pkgModel.InvalidTokenError(err.Error()),
			}
		}
	}

	mt, err := mytoken.ParseJWT(string(req.Mytoken))
	if err != nil {
		return &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: pkgModel.InvalidTokenError(err.Error()),
		}
	}
	log.Trace("Parsed mytoken")

	revoked, dbErr := dbhelper.CheckTokenRevoked(mt.ID)
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

	if ok := mt.VerifyCapabilities(capabilities.CapabilityCreateMT); !ok {
		return &model.Response{
			Status:   fiber.StatusForbidden,
			Response: pkgModel.APIErrorInsufficientCapabilities,
		}
	}
	log.Trace("Checked mytoken capabilities")
	if ok := mt.Restrictions.VerifyForOther(nil, ctx.IP(), mt.ID); !ok {
		return &model.Response{
			Status:   fiber.StatusForbidden,
			Response: pkgModel.APIErrorUsageRestricted,
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
	req.Restrictions.ReplaceThisIp(ctx.IP())
	return handleMytokenFromMytoken(mt, req, ctxUtils.ClientMetaData(ctx), req.ResponseType)
}

func handleMytokenFromMytoken(parent *mytoken.Mytoken, req *response.MytokenFromMytokenRequest, networkData *model.ClientMetaData, responseType pkgModel.ResponseType) *model.Response {
	ste, errorResponse := createMytokenEntry(parent, req, *networkData)
	if errorResponse != nil {
		return errorResponse
	}
	if err := db.Transact(func(tx *sqlx.Tx) error {
		if len(parent.Restrictions) > 0 {
			if err := parent.Restrictions.GetValidForOther(tx, networkData.IP, parent.ID)[0].UsedOther(tx, parent.ID); err != nil {
				return err
			}
		}
		if err := ste.Store(tx, "Used grant_type mytoken"); err != nil {
			return err
		}
		return eventService.LogEvents(tx, []eventService.MTEvent{
			{event.FromNumber(event.MTEventInheritedRT, "Got RT from parent"), ste.ID},
			{event.FromNumber(event.MTEventMTCreated, strings.TrimSpace(fmt.Sprintf("Created MT %s", req.Name))), parent.ID},
		}, *networkData)
	}); err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}

	res, err := ste.Token.ToTokenResponse(responseType, *networkData, "")
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: res,
	}
}

func createMytokenEntry(parent *mytoken.Mytoken, req *response.MytokenFromMytokenRequest, networkData model.ClientMetaData) (*mytokenrepo.MytokenEntry, *model.Response) {
	rtID, dbErr := refreshtokenrepo.GetRTID(nil, parent.ID)
	rtFound, err := dbhelper.ParseError(dbErr)
	if err != nil {
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
	r := restrictions.Tighten(parent.Restrictions, req.Restrictions)
	capsFromParent := parent.SubtokenCapabilities
	if capsFromParent == nil {
		capsFromParent = parent.Capabilities
	}
	c := capabilities.Tighten(capsFromParent, req.Capabilities)
	if len(c) == 0 {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: pkgModel.BadRequestError("mytoken to be issued cannot have any of the requested capabilities"),
		}
	}
	var sc capabilities.Capabilities = nil
	if c.Has(capabilities.CapabilityCreateMT) {
		sc = capabilities.Tighten(capsFromParent, req.SubtokenCapabilities)
	}
	ste := mytokenrepo.NewMytokenEntry(
		mytoken.NewMytoken(parent.OIDCSubject, parent.OIDCIssuer, r, c, sc),
		req.Name, networkData)
	encryptionKey, _, err := refreshtokenrepo.GetEncryptionKey(nil, parent.ID, string(req.Mytoken))
	if err != nil {
		return ste, model.ErrorToInternalServerErrorResponse(err)
	}
	if err = ste.SetRefreshToken(rtID, encryptionKey); err != nil {
		return ste, model.ErrorToInternalServerErrorResponse(err)
	}
	ste.ParentID = parent.ID
	ste.RootID = rootID
	return ste, nil
}

// RevokeMytoken revokes a Mytoken
func RevokeMytoken(tx *sqlx.Tx, id mtid.MTID, token token.Token, recursive bool, issuer string) *model.Response {
	provider, ok := config.Get().ProviderByIssuer[issuer]
	if !ok {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: pkgModel.APIErrorUnknownIssuer,
		}
	}
	if err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		rtID, err := refreshtokenrepo.GetRTID(tx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}
		rt, _, err := refreshtokenrepo.GetRefreshToken(tx, id, string(token))
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
			apiError := e.Response.(pkgModel.APIError)
			return fmt.Errorf("%s: %s", apiError.Error, apiError.ErrorDescription)
		}
		return refreshtokenrepo.DeleteRefreshToken(tx, rtID)
	}); err != nil {
		if strings.HasPrefix(err.Error(), "oidc_error") {
			return &model.Response{
				Status:   httpStatus.StatusOIDPError,
				Response: pkgModel.OIDCError(err.Error(), ""),
			}
		} else {
			return model.ErrorToInternalServerErrorResponse(err)
		}
	}
	return nil
}
