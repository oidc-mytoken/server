package supertoken

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/supertokenrepo"
	dbhelper "github.com/oidc-mytoken/server/internal/db/dbrepo/supertokenrepo/supertokenrepohelper"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/supertokenrepo/transfercoderepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/super/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/revoke"
	"github.com/oidc-mytoken/server/internal/server/httpStatus"
	"github.com/oidc-mytoken/server/internal/utils/ctxUtils"
	pkgModel "github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/supertoken/capabilities"
	eventService "github.com/oidc-mytoken/server/shared/supertoken/event"
	event "github.com/oidc-mytoken/server/shared/supertoken/event/pkg"
	supertoken "github.com/oidc-mytoken/server/shared/supertoken/pkg"
	"github.com/oidc-mytoken/server/shared/supertoken/pkg/stid"
	"github.com/oidc-mytoken/server/shared/supertoken/restrictions"
	"github.com/oidc-mytoken/server/shared/supertoken/token"
	"github.com/oidc-mytoken/server/shared/utils"
)

// HandleSuperTokenFromTransferCode handles requests to return the super token for a transfer code
func HandleSuperTokenFromTransferCode(ctx *fiber.Ctx) *model.Response {
	log.Debug("Handle supertoken from transfercode")
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
		tokenStr, err = transfercoderepo.PopTokenForTransferCode(tx, req.TransferCode)
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
	jwt, err := token.GetLongSuperToken(tokenStr)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	st, err := supertoken.ParseJWT(string(jwt))
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{
		Status: fiber.StatusOK,
		Response: response.SuperTokenResponse{
			SuperToken:           tokenStr,
			SuperTokenType:       tokenType,
			ExpiresIn:            st.ExpiresIn(),
			Restrictions:         st.Restrictions,
			Capabilities:         st.Capabilities,
			SubtokenCapabilities: st.SubtokenCapabilities,
		},
	}

}

// HandleSuperTokenFromSuperToken handles requests to create a super token from an existing super token
func HandleSuperTokenFromSuperToken(ctx *fiber.Ctx) *model.Response {
	log.Debug("Handle supertoken from supertoken")
	req := response.NewSuperTokenRequest()
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	log.Trace("Parsed super token request")

	// GrantType already checked

	if len(req.SuperToken) == 0 {
		var err error
		req.SuperToken, err = token.GetLongSuperToken(ctx.Cookies("mytoken-supertoken"))
		if err != nil {
			return &model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: pkgModel.InvalidTokenError(err.Error()),
			}
		}
	}

	st, err := supertoken.ParseJWT(string(req.SuperToken))
	if err != nil {
		return &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: pkgModel.InvalidTokenError(err.Error()),
		}
	}
	log.Trace("Parsed super token")

	revoked, dbErr := dbhelper.CheckTokenRevoked(st.ID)
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

	if ok := st.VerifyCapabilities(capabilities.CapabilityCreateST); !ok {
		return &model.Response{
			Status:   fiber.StatusForbidden,
			Response: pkgModel.APIErrorInsufficientCapabilities,
		}
	}
	log.Trace("Checked super token capabilities")
	if ok := st.Restrictions.VerifyForOther(nil, ctx.IP(), st.ID); !ok {
		return &model.Response{
			Status:   fiber.StatusForbidden,
			Response: pkgModel.APIErrorUsageRestricted,
		}
	}
	log.Trace("Checked super token restrictions")

	if req.Issuer == "" {
		req.Issuer = st.OIDCIssuer
	} else {
		if req.Issuer != st.OIDCIssuer {
			return &model.Response{
				Status:   fiber.StatusBadRequest,
				Response: pkgModel.BadRequestError("token not for specified issuer"),
			}
		}
		log.Trace("Checked issuer")
	}
	req.Restrictions.ReplaceThisIp(ctx.IP())
	return handleSuperTokenFromSuperToken(st, req, ctxUtils.ClientMetaData(ctx), req.ResponseType)
}

func handleSuperTokenFromSuperToken(parent *supertoken.SuperToken, req *response.SuperTokenFromSuperTokenRequest, networkData *model.ClientMetaData, responseType pkgModel.ResponseType) *model.Response {
	ste, errorResponse := createSuperTokenEntry(parent, req, *networkData)
	if errorResponse != nil {
		return errorResponse
	}
	if err := db.Transact(func(tx *sqlx.Tx) error {
		if len(parent.Restrictions) > 0 {
			if err := parent.Restrictions.GetValidForOther(tx, networkData.IP, parent.ID)[0].UsedOther(tx, parent.ID); err != nil {
				return err
			}
		}
		if err := ste.Store(tx, "Used grant_type super_token"); err != nil {
			return err
		}
		return eventService.LogEvent(tx, event.FromNumber(event.STEventInheritedRT, "Got RT from parent"), ste.ID, *networkData)
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

func createSuperTokenEntry(parent *supertoken.SuperToken, req *response.SuperTokenFromSuperTokenRequest, networkData model.ClientMetaData) (*supertokenrepo.SuperTokenEntry, *model.Response) {
	rt, rtFound, dbErr := dbhelper.GetRefreshToken(parent.ID, string(req.SuperToken))
	if dbErr != nil {
		return nil, model.ErrorToInternalServerErrorResponse(dbErr)
	}
	if !rtFound {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: pkgModel.InvalidTokenError(""),
		}
	}
	rootID, rootFound, dbErr := dbhelper.GetSTRootID(parent.ID)
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
	var sc capabilities.Capabilities = nil
	if c.Has(capabilities.CapabilityCreateST) {
		sc = capabilities.Tighten(capsFromParent, req.SubtokenCapabilities)
	}
	ste := supertokenrepo.NewSuperTokenEntry(
		supertoken.NewSuperToken(parent.OIDCSubject, parent.OIDCIssuer, r, c, sc),
		req.Name, networkData)
	encryptionKey, _, err := dbhelper.GetEncryptionKey(nil, parent.ID, string(req.SuperToken))
	if err != nil {
		return ste, model.ErrorToInternalServerErrorResponse(err)
	}
	if err = ste.SetRefreshToken(rt, encryptionKey); err != nil {
		return ste, model.ErrorToInternalServerErrorResponse(err)
	}
	ste.ParentID = parent.ID
	ste.RootID = rootID
	return ste, nil
}

// RevokeSuperToken revokes a super token
func RevokeSuperToken(tx *sqlx.Tx, id stid.STID, token token.Token, recursive bool, issuer string) *model.Response {
	rt, rtFound, dbErr := dbhelper.GetRefreshToken(id, string(token))
	if dbErr != nil {
		return model.ErrorToInternalServerErrorResponse(dbErr)
	}
	if !rtFound {
		return nil
	}
	provider, ok := config.Get().ProviderByIssuer[issuer]
	if !ok {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: pkgModel.APIErrorUnknownIssuer,
		}
	}
	if err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		if err := dbhelper.RevokeST(tx, id, recursive); err != nil {
			return err
		}
		count, err := dbhelper.CountRTOccurrences(tx, rt)
		if err != nil {
			return err
		}
		if count == 0 {
			if e := revoke.RefreshToken(provider, rt); e != nil {
				apiError := e.Response.(pkgModel.APIError)
				return fmt.Errorf("%s: %s", apiError.Error, apiError.ErrorDescription)
			}
		}
		return nil
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
