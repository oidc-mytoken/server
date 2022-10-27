package cookies

import (
	"net/url"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/server/apipath"
)

const (
	mytokenCookieName      = "mytoken"
	transferCodeCookieName = "mytoken-transfercode"
)

var cookieDomain string
var cookieSecure bool

// CookieLifetime is the lifetime of a cookie
var CookieLifetime int

// Init initializes cookie values
func Init() {
	cookieSecure = config.Get().Server.Secure
	CookieLifetime = config.Get().Features.OIDCFlows.AuthCode.Web.CookieLifetime
	u, err := url.Parse(config.Get().IssuerURL)
	if err != nil {
		log.WithError(err).Error("Cannot get domain from issuer url")
	}
	cookieDomain = u.Hostname()
}

// MytokenCookie creates a fiber.Cookie for the passed mytoken
func MytokenCookie(mytoken string) *fiber.Cookie {
	return &fiber.Cookie{
		Name:     mytokenCookieName,
		Value:    mytoken,
		Path:     apipath.Prefix,
		Domain:   cookieDomain,
		MaxAge:   CookieLifetime,
		Secure:   cookieSecure,
		HTTPOnly: true,
		SameSite: "Strict",
	}
}

// TransferCodeCookie creates a fiber.Cookie for the passed transfer code
func TransferCodeCookie(transferCode string, expiresIn int) *fiber.Cookie {
	return &fiber.Cookie{
		Name:     transferCodeCookieName,
		Value:    transferCode,
		Path:     apipath.Prefix,
		Domain:   cookieDomain,
		MaxAge:   expiresIn,
		Secure:   cookieSecure,
		HTTPOnly: true,
		SameSite: "Strict",
	}
}
