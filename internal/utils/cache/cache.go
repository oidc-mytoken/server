package cache

import (
	"fmt"
	"time"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/model"
)

type Cache interface {
	Get(key string, target any) (bool, error)
	Set(key string, value any, expiration time.Duration) error
}

var c Cache

// SetCache sets the cache
func SetCache(cache Cache) {
	c = cache
}

// InitCache initializes the cache according to the configuration
func InitCache() {
	if c != nil {
		return
	}
	if config.Get().Caching.External != nil {
		if config.Get().Caching.External.Redis != nil {
			initRedisCache()
		}
	}
	if c == nil {
		initInternalCache()
	}
}

// Type defines the type of cache
type Type int

// Different cache types
const (
	IPHostCache Type = iota
	invalidated1
	invalidated2
	WebProfiles
	FederationLib
	FederationOPMetadata
	ScheduledNotifications
	IPCache
)

func k(t Type, key string) string {
	return fmt.Sprintf("%d:%s", t, key)
}

// Set sets a value in a cache
func Set(
	t Type, key string, value interface{},
	expiration ...time.Duration,
) error {
	exp := time.Duration(0)
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	return c.Set(k(t, key), value, exp)
}

// Get returns the cached value for a given key
func Get(t Type, key string, i any) (bool, error) {
	return c.Get(k(t, key), i)
}

type subcache struct {
	t Type
}

// Get implements the Cache interface
func (sc subcache) Get(key string, i any) (bool, error) {
	return Get(sc.t, key, i)
}

// Set implements the Cache interface
func (sc subcache) Set(key string, value any, expiration time.Duration) error {
	return Set(sc.t, key, value, expiration)
}

// SubCache returns a sub-cache for the given Type
func SubCache(t Type) Cache {
	return subcache{t}
}

// SetIPParseResult caches the result of an ip parsing
func SetIPParseResult(ipStr string, ip model.IPParseResult) error {
	return Set(IPCache, ipStr, &ip)
}

// GetIPParseResult returns the cached result of an ip parsing
func GetIPParseResult(ipStr string) (
	result model.IPParseResult, found bool,
	err error,
) {
	found, err = Get(IPCache, ipStr, &result)
	return
}
