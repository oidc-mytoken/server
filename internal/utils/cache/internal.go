package cache

import (
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/oidc-mytoken/server/internal/config"
)

func initInternalCache() {
	SetCache(
		cache.New(
			time.Duration(config.Get().Caching.Internal.DefaultExpiration)*time.Second,
			time.Duration(config.Get().Caching.Internal.CleanupInterval)*time.Second,
		),
	)
}
