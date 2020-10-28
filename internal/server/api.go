package server

import "github.com/gofiber/fiber/v2"

func addAPIRoutes(s fiber.Router) {
	api := s.Group("/api")
	addAPIv0Routes(api)
}

func addAPIv0Routes(s fiber.Router) {
	api := s.Group("/v0")
	tokens := api.Group("/token")
	tokens.Get("/super")
	api.Get("/something")
}
