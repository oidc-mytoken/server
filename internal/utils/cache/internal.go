package cache

import (
	"time"

	"github.com/TwiN/gocache/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/oidc-mytoken/server/internal/config"
)

func initInternalCache() {
	SetCache(
		newCacheWrapper(time.Duration(config.Get().Caching.Internal.DefaultExpiration) * time.Second),
	)
}

type cacheWrapper struct {
	c *gocache.Cache
}

func NewInternalCache(defaultExpiration time.Duration) Cache {
	return newCacheWrapper(defaultExpiration)
}

func newCacheWrapper(defaultExpiration time.Duration) cacheWrapper {
	cache := gocache.NewCache().WithDefaultTTL(defaultExpiration)
	if err := cache.StartJanitor(); err != nil {
		log.WithError(err).Fatal("could not init cache")
	}
	return cacheWrapper{
		cache,
	}
}

func (c cacheWrapper) Get(key string, target any) (bool, error) {
	entryV, ok := c.c.Get(key)
	if !ok {
		return false, nil
	}
	entry, ok := entryV.([]byte)
	if !ok {
		log.Error("invalid cache entry type")
		return false, errors.New("invalid cache entry type")
	}
	return true, msgpack.Unmarshal(entry, target)
}

func (c cacheWrapper) Set(key string, value any, expiration time.Duration) error {
	data, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}
	c.c.SetWithTTL(key, data, expiration)
	return nil
}
