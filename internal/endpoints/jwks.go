package endpoints

import (
	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/jws"
)

// HandleJWKS handles request for the jwks, returning the jwks
func HandleJWKS(ctx *fiber.Ctx) error {
	return ctx.JSON(jws.GetJWKS(jws.KeyUsageMytokenSigning))
}
