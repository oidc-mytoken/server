package server

import (
	"github.com/gofiber/fiber/v2"

	"github.com/zachmann/mytoken/internal/model/version"

	"github.com/zachmann/mytoken/internal/server/config"
	"github.com/zachmann/mytoken/internal/server/endpoints/revocation"
	"github.com/zachmann/mytoken/internal/server/endpoints/token/access"
	"github.com/zachmann/mytoken/internal/server/endpoints/token/super"
	"github.com/zachmann/mytoken/internal/server/server/routes"
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
}
