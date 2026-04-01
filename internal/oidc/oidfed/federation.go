package oidfed

import (
	oidfedcache "github.com/go-oidfed/lib/cache"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/endpoints/federation"
	"github.com/oidc-mytoken/server/internal/jws"
	"github.com/oidc-mytoken/server/internal/utils/cache"
)

// Init inits the oidfed
func Init() {
	if !config.Get().Features.Federation.Enabled {
		return
	}
	jws.LoadFederationKey()
	jws.LoadOIDCSigningKey()
	oidfedcache.SetCache(cache.SubCache(cache.FederationLib))
	federation.InitEntityConfiguration()
	Discovery()
}
