package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/endpoints/revocation"
	"github.com/zachmann/mytoken/internal/endpoints/token/access"
	"github.com/zachmann/mytoken/internal/endpoints/token/super"
)

func addAPIRoutes(s fiber.Router) {
	api := s.Group("/api")
	addAPIv0Routes(api)
}

func addAPIv0Routes(s fiber.Router) {
	api := s.Group("/v0")
	tokens := api.Group("/token")
	tokens.Post("/super", super.HandleSuperTokenEndpoint)
	tokens.Post("/access", access.HandleAccessTokenEndpoint)
	if config.Get().Features.TokenRevocation.Enabled {
		tokens.Post("/revoke", revocation.HandleRevoke)
	}
}
