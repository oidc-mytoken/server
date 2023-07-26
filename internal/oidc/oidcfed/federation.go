package oidcfed

import (
	oidcfedcache "github.com/zachmann/go-oidcfed/pkg/cache"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/endpoints/federation"
	"github.com/oidc-mytoken/server/internal/jws"
	"github.com/oidc-mytoken/server/internal/utils/cache"
)

func Init() {
	if !config.Get().Features.Federation.Enabled {
		return
	}
	jws.LoadFederationKey()
	jws.LoadOIDCSigningKey()
	oidcfedcache.SetCache(cache.SubCache(cache.FederationLib))
	Discovery()
	federation.InitEntityConfiguration()
}
