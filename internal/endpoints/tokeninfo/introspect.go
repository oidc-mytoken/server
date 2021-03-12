package tokeninfo

import (
	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	pkgModel "github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/supertoken/capabilities"
	supertoken "github.com/oidc-mytoken/server/shared/supertoken/pkg"
)

func handleTokenInfoIntrospect(st *supertoken.SuperToken) model.Response {
	// If we call this function it means the token is valid.

	if !st.Capabilities.Has(capabilities.CapabilityTokeninfoIntrospect) {
		return model.Response{
			Status:   fiber.StatusForbidden,
			Response: pkgModel.APIErrorInsufficientCapabilities,
		}
	}
	usedToken, err := st.ToUsedSuperToken()
	if err != nil {
		return *model.ErrorToInternalServerErrorResponse(err)
	}
	return model.Response{
		Status: fiber.StatusOK,
		Response: pkg.TokeninfoIntrospectResponse{
			Valid: true,
			Token: *usedToken,
		},
	}
}
