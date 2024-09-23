package oidcfed

import (
	oidfedcache "github.com/zachmann/go-oidfed/pkg/cache"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/endpoints/federation"
	"github.com/oidc-mytoken/server/internal/jws"
	"github.com/oidc-mytoken/server/internal/utils/cache"
)

// Init inits the oidcfed
func Init() {
	if !config.Get().Features.Federation.Enabled {
		return
	}
	jws.LoadFederationKey()
	jws.LoadOIDCSigningKey()
	oidfedcache.SetCache(cache.SubCache(cache.FederationLib))
	Discovery()
	federation.InitEntityConfiguration()
}
