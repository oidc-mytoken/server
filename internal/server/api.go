package server

import (
	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/endpoints/revocation"
	"github.com/oidc-mytoken/server/internal/endpoints/token/access"
	"github.com/oidc-mytoken/server/internal/endpoints/token/super"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo"
	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/server/routes"
)

func addAPIRoutes(s fiber.Router) {
	for v := config.Get().API.MinVersion; v <= version.MAJOR; v++ {
		addAPIvXRoutes(s, v)
	}
}

func addAPIvXRoutes(s fiber.Router, version int) {
	apiPaths := routes.GetAPIPaths(version)
	s.Post(apiPaths.SuperTokenEndpoint, super.HandleSuperTokenEndpoint)
	s.Post(apiPaths.AccessTokenEndpoint, access.HandleAccessTokenEndpoint)
	if config.Get().Features.TokenRevocation.Enabled {
		s.Post(apiPaths.RevocationEndpoint, revocation.HandleRevoke)
	}
	if config.Get().Features.TransferCodes.Enabled {
		s.Post(apiPaths.TokenTransferEndpoint, super.HandleCreateTransferCodeForExistingSuperToken)
	}
	if config.Get().Features.TokenInfo.Enabled {
		s.Post(apiPaths.TokenInfoEndpoint, tokeninfo.HandleTokenInfo)
	}
}
