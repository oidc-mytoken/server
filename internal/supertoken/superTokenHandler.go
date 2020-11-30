package supertoken

import (
	"encoding/json"

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
