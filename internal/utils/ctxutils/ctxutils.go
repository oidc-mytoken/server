package ctxutils

import (
	"net/url"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

// Params is a wrapper around ctx.Params that also url decodes the value
func Params(ctx *fiber.Ctx, key string, defaultValue ...string) string {
	v := ctx.Params(key, defaultValue...)
	decodedString, err := url.QueryUnescape(v)
	if err != nil {
		log.WithError(err).Error("Failed to unescape params")
		return v
	}
	return decodedString
}
