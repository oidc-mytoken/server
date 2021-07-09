package cookies

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/server/internal/config"
)

const (
	cookieAge              = 3600 * 24 * 7 //TODO from config, same as in js
	mytokenCookieName      = "mytoken"
	transferCodeCookieName = "mytoken-transfercode"
)

// MytokenCookie creates a fiber.Cookie for the passed mytoken
func MytokenCookie(mytoken string) fiber.Cookie {
	return fiber.Cookie{
		Name:     mytokenCookieName,
		Value:    mytoken,
		Path:     "/api",
		MaxAge:   cookieAge,
		Secure:   config.Get().Server.Secure,
		HTTPOnly: true,
		SameSite: "Strict",
	}
}

// TransferCodeCookie creates a fiber.Cookie for the passed transfer code
func TransferCodeCookie(transferCode string, expiresIn int) fiber.Cookie {
	return fiber.Cookie{
		Name:     transferCodeCookieName,
		Value:    transferCode,
		Path:     "/api",
		MaxAge:   expiresIn,
		Secure:   config.Get().Server.Secure,
		HTTPOnly: true,
		SameSite: "Strict",
	}
}
