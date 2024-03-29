package cache

import (
	"fmt"
	"net"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/oidc-mytoken/server/internal/config"
)

type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any, expiration time.Duration)
}

var c Cache

func setCache(cache Cache) {
	c = cache
}

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
	IPAddrCache
	IPNetCache
	WebProfiles
	FederationLib
	FederationOPMetadata
)

func k(t Type, key string) string {
	return fmt.Sprintf("%d:%s", t, key)
}

// Set sets a value in a cache
func Set(t Type, key string, value interface{}, expiration ...time.Duration) {
	exp := cache.DefaultExpiration
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	c.Set(k(t, key), value, exp)
}

// Get returns the cached value for a given key
func Get(t Type, key string) (interface{}, bool) {
	return c.Get(k(t, key))
}

type subcache struct {
	t Type
}

func (sc subcache) Get(key string) (any, bool) {
	return Get(sc.t, key)
}

func (sc subcache) Set(key string, value any, expiration time.Duration) {
	Set(sc.t, key, value, expiration)
}

func SubCache(t Type) Cache {
	return subcache{t}
}

// SetIPParseResult caches the result of an ip parsing
func SetIPParseResult(ipStr string, ip net.IP, ipNet *net.IPNet) {
	Set(IPAddrCache, ipStr, ip)
	Set(IPNetCache, ipStr, ipNet)
}

// GetIPParseResult returns the cached result of an ip parsing
func GetIPParseResult(ipStr string) (net.IP, *net.IPNet, bool) {
	ip, ipFound := Get(IPAddrCache, ipStr)
	if !ipFound {
		return nil, nil, false
	}
	ipNet, netFound := Get(IPNetCache, ipStr)
	if !netFound {
		return nil, nil, false
	}
	return ip.(net.IP), ipNet.(*net.IPNet), true
}
