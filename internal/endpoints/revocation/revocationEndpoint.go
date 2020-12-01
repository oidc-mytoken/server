package revocation

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	request "github.com/zachmann/mytoken/internal/endpoints/revocation/pkg"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/supertoken"
	supertokenPkg "github.com/zachmann/mytoken/internal/supertoken/pkg"
)

func HandleRevoke(ctx *fiber.Ctx) (err error) {
	log.Debug("Handle revocation request")
	req := request.RevocationRequest{}
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	log.Trace("Parsed super token request")
	if req.OIDCIssuer == "" {
		res := &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError("required parameter 'oidc_issuer' is missing"),
		}
		return res.Send(ctx)
	}
	switch len(req.Token) {
	case 0:
		return ctx.SendStatus(fiber.StatusNoContent)
	//case TransferCodeLen: err = revokeTransferCode(req.Token, req.Recursive)
	//case ShortSuperToken: err = revokeShortSuperToken(req.Token, req.Recursive)
	default:
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
		errorRes := supertoken.RevokeSuperToken(req.Token, req.Recursive, req.OIDCIssuer)
		if errorRes != nil {
			return errorRes.Send(ctx)
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	}
}
