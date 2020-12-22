package revocation

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/db/dbrepo/supertokenrepo/shorttokenrepo"
	request "github.com/zachmann/mytoken/internal/endpoints/revocation/pkg"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/supertoken"
	supertokenPkg "github.com/zachmann/mytoken/internal/supertoken/pkg"
	"github.com/zachmann/mytoken/internal/supertoken/token"
	"github.com/zachmann/mytoken/internal/utils"
)

// HandleRevoke handles requests to the revocation endpoint
func HandleRevoke(ctx *fiber.Ctx) error {
	log.Debug("Handle revocation request")
	req := request.RevocationRequest{}
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	log.Trace("Parsed super token request")
	if req.OIDCIssuer == "" {
		return model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError("required parameter 'oidc_issuer' is missing"),
		}.Send(ctx)
	}
	if utils.IsJWT(req.Token) { // normal SuperToken
		return revokeSuperToken(ctx, req)
	} else if len(req.Token) == config.Get().Features.TransferCodes.Len { // Transfer Code
		return model.ResponseNYI.Send(ctx) //TODO
	} else { // Short Token
		shortToken, found, err := shorttokenrepo.ParseShortToken(nil, req.Token)
		if err != nil {
			return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
		}
		if !found {
			return ctx.SendStatus(fiber.StatusNoContent)
		}
		if err = shortToken.Delete(nil); err != nil {
			return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
		}
		jwt, err := shortToken.JWT()
		if err != nil {
			return ctx.SendStatus(fiber.StatusNoContent)
		}
		req.Token = jwt
		return revokeSuperToken(ctx, req)
	}
}

func revokeSuperToken(ctx *fiber.Ctx, req request.RevocationRequest) error {
	st, err := supertokenPkg.ParseJWT(req.Token)
	if err != nil {
		return ctx.SendStatus(fiber.StatusNoContent)
	}
	if st.OIDCIssuer != req.OIDCIssuer {
		res := &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError("token not for specified issuer"),
		}
		return res.Send(ctx)
	}
	errorRes := supertoken.RevokeSuperToken(st.ID, token.Token(req.Token), req.Recursive, req.OIDCIssuer)
	if errorRes != nil {
		return errorRes.Send(ctx)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
