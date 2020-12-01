package supertoken

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zachmann/mytoken/internal/server/httpStatus"

	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/oidc/revoke"

	"github.com/jmoiron/sqlx"
	"github.com/zachmann/mytoken/internal/db"

	eventService "github.com/zachmann/mytoken/internal/supertoken/event"
	event "github.com/zachmann/mytoken/internal/supertoken/event/pkg"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"github.com/zachmann/mytoken/internal/db/dbModels"
	response "github.com/zachmann/mytoken/internal/endpoints/token/super/pkg"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/supertoken/capabilities"
	supertoken "github.com/zachmann/mytoken/internal/supertoken/pkg"
	"github.com/zachmann/mytoken/internal/supertoken/restrictions"
	"github.com/zachmann/mytoken/internal/utils/ctxUtils"
	"github.com/zachmann/mytoken/internal/utils/dbUtils"
)

func HandleSuperTokenFromSuperToken(ctx *fiber.Ctx) *model.Response {
	log.Debug("Handle supertoken from supertoken")
	req := response.NewSuperTokenRequest()
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	log.Trace("Parsed super token request")

	// GrantType already checked

	revoked, dbErr := dbUtils.CheckTokenRevoked(req.SuperToken)
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

	st, err := supertoken.ParseJWT(req.SuperToken)
	if err != nil {
		return &model.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.InvalidTokenError(err.Error()),
		}
	}
	log.Trace("Parsed super token")
	if ok := st.VerifyCapabilities(capabilities.CapabilityCreateST); !ok {
		return &model.Response{
			Status:   fiber.StatusForbidden,
			Response: model.APIErrorInsufficientCapabilities,
		}
	}
	log.Trace("Checked super token capabilities")
	if ok := st.Restrictions.VerifyForOther(ctx.IP(), st.ID); !ok {
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

	return handleSuperTokenFromSuperToken(st, req, ctxUtils.NetworkData(ctx))
}

func handleSuperTokenFromSuperToken(parent *supertoken.SuperToken, req *response.SuperTokenFromSuperTokenRequest, networkData *model.NetworkData) *model.Response {
	ste, errorResponse := createSuperTokenEntry(parent, req, *networkData)
	if errorResponse != nil {
		return errorResponse
	}
	if err := parent.Restrictions.GetValidForOther(networkData.IP, parent.ID)[0].UsedOther(parent.ID); err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if err := ste.Store("Used grant_type super_token"); err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if err := eventService.LogEvent(event.FromNumber(event.STEventInheritedRT, "Got RT from parent"), ste.ID, *networkData); err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}

	return &model.Response{
		Status:   fiber.StatusOK,
		Response: ste.Token.ToSuperTokenResponse(""),
	}
}

func createSuperTokenEntry(parent *supertoken.SuperToken, req *response.SuperTokenFromSuperTokenRequest, networkData model.NetworkData) (*dbModels.SuperTokenEntry, *model.Response) {
	rt, rtFound, dbErr := dbUtils.GetRefreshToken(parent.ID)
	if dbErr != nil {
		return nil, model.ErrorToInternalServerErrorResponse(dbErr)
	}
	if !rtFound {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.InvalidTokenError("token unknown"),
		}
	}
	rootID, rootFound, dbErr := dbUtils.GetSTRootID(parent.ID)
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
	ste := dbModels.NewSuperTokenEntry(req.Name, parent.OIDCSubject, parent.OIDCIssuer, r, c, sc, networkData)
	ste.RefreshToken = rt
	ste.ParentID = parent.ID.String()
	ste.RootID = rootID
	return ste, nil
}

func RevokeSuperToken(token string, recursive bool, issuer string) *model.Response {
	rt, rtFound, dbErr := dbUtils.GetRefreshTokenByTokenString(token)
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
	if err := db.Transact(func(tx *sqlx.Tx) error {
		if recursive {
			if err := dbUtils.RecursiveRevokeSTByTokenString(token, tx); err != nil {
				return err
			}
		} else {
			if _, err := tx.Exec(`DELETE FROM SuperTokens WHERE token=?`, token); err != nil {
				return err
			}
		}
		var count int
		if err := tx.Get(&count, `SELECT COUNT(1) FROM SuperTokens WHERE refresh_token=?`, rt); err != nil {
			return err
		}
		if count == 0 {
			if err := revoke.RevokeRefreshToken(provider, rt); err != nil {
				e := err.Response.(model.APIError)
				return fmt.Errorf("%s: %s", e.Error, e.ErrorDescription)
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
