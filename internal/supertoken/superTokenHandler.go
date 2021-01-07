package supertoken

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/db/dbrepo/supertokenrepo"
	dbhelper "github.com/zachmann/mytoken/internal/db/dbrepo/supertokenrepo/supertokenrepohelper"
	"github.com/zachmann/mytoken/internal/db/dbrepo/supertokenrepo/transfercoderepo"
	response "github.com/zachmann/mytoken/internal/endpoints/token/super/pkg"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/oidc/revoke"
	"github.com/zachmann/mytoken/internal/server/httpStatus"
	"github.com/zachmann/mytoken/internal/supertoken/capabilities"
	eventService "github.com/zachmann/mytoken/internal/supertoken/event"
	event "github.com/zachmann/mytoken/internal/supertoken/event/pkg"
	supertoken "github.com/zachmann/mytoken/internal/supertoken/pkg"
	"github.com/zachmann/mytoken/internal/supertoken/restrictions"
	"github.com/zachmann/mytoken/internal/supertoken/token"
	"github.com/zachmann/mytoken/internal/utils"
	"github.com/zachmann/mytoken/internal/utils/ctxUtils"
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
				Response: model.APIErrorBadTransferCode,
			}
			return fmt.Errorf("error_res")
		}
		if status.Expired {
			errorRes = &model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: model.APIErrorTransferCodeExpired,
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

	tokenType := model.ResponseTypeToken
	if !utils.IsJWT(tokenStr) {
		tokenType = model.ResponseTypeShortToken
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

	st, err := supertoken.ParseJWT(string(req.SuperToken))
	if err != nil {
		return &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.InvalidTokenError(err.Error()),
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
			Response: model.InvalidTokenError("not a valid token"),
		}
	}
	log.Trace("Checked token not revoked")

	if ok := st.VerifyCapabilities(capabilities.CapabilityCreateST); !ok {
		return &model.Response{
			Status:   fiber.StatusForbidden,
			Response: model.APIErrorInsufficientCapabilities,
		}
	}
	log.Trace("Checked super token capabilities")
	if ok := st.Restrictions.VerifyForOther(nil, ctx.IP(), st.ID); !ok {
		return &model.Response{
			Status:   fiber.StatusForbidden,
			Response: model.APIErrorUsageRestricted,
		}
	}
	log.Trace("Checked super token restrictions")

	if req.Issuer == "" {
		req.Issuer = st.OIDCIssuer
	} else {
		if req.Issuer != st.OIDCIssuer {
			return &model.Response{
				Status:   fiber.StatusBadRequest,
				Response: model.BadRequestError("token not for specified issuer"),
			}
		}
		log.Trace("Checked issuer")
	}

	return handleSuperTokenFromSuperToken(st, req, ctxUtils.ClientMetaData(ctx), req.ResponseType)
}

func handleSuperTokenFromSuperToken(parent *supertoken.SuperToken, req *response.SuperTokenFromSuperTokenRequest, networkData *model.ClientMetaData, responseType model.ResponseType) *model.Response {
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
		if err := eventService.LogEvent(tx, event.FromNumber(event.STEventInheritedRT, "Got RT from parent"), ste.ID, *networkData); err != nil {
			return err
		}
		return nil
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
	rt, rtFound, dbErr := dbhelper.GetRefreshToken(parent.ID, req.SuperToken)
	if dbErr != nil {
		return nil, model.ErrorToInternalServerErrorResponse(dbErr)
	}
	if !rtFound {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.InvalidTokenError("token unknown"),
		}
	}
	rootID, rootFound, dbErr := dbhelper.GetSTRootID(parent.ID)
	if dbErr != nil {
		return nil, model.ErrorToInternalServerErrorResponse(dbErr)
	}
	if !rootFound {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.InvalidTokenError("token unknown"),
		}
	}
	if rootID == "" {
		rootID = parent.ID.String()
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
	ste.RefreshToken = rt
	ste.ParentID = parent.ID.String()
	ste.RootID = rootID
	return ste, nil
}

// RevokeSuperToken revokes a super token
func RevokeSuperToken(tx *sqlx.Tx, id uuid.UUID, token token.Token, recursive bool, issuer string) *model.Response {
	rt, rtFound, dbErr := dbhelper.GetRefreshToken(id, token)
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
			Response: model.APIErrorUnknownIssuer,
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
				apiError := e.Response.(model.APIError)
				return fmt.Errorf("%s: %s", apiError.Error, apiError.ErrorDescription)
			}
		}
		return nil
	}); err != nil {
		if strings.HasPrefix(err.Error(), "oidc_error") {
			return &model.Response{
				Status:   httpStatus.StatusOIDPError,
				Response: model.OIDCError(err.Error(), ""),
			}
		} else {
			return model.ErrorToInternalServerErrorResponse(err)
		}
	}
	return nil
}
