package endpoints

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zachmann/mytoken/internal/jws"
)

func HandleJWKS(ctx *fiber.Ctx) error {
	return ctx.JSON(jws.GetJWKS())
}
