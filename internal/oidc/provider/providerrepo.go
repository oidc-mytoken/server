package provider

import (
	"github.com/oidc-mytoken/utils/utils/issuerutils"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/oidcfed"
)

var fileProviderByIssuer map[string]model.Provider

// Init inits the configured providers
func Init() {
	fileProviderByIssuer = make(map[string]model.Provider)
	for _, p := range config.Get().Providers {
		iss0, iss1 := issuerutils.GetIssuerWithAndWithoutSlash(p.Issuer)
		fileProviderByIssuer[iss0] = SimpleProvider{p}
		fileProviderByIssuer[iss1] = SimpleProvider{p}
	}
}

// GetProvider returns the model.Provider for a passed issuer
func GetProvider(issuer string) model.Provider {
	if p, ok := fileProviderByIssuer[issuer]; ok {
		return p
	}
	if config.Get().Features.Federation.Enabled {
		return oidcfed.GetOIDCFedProvider(issuer)
	}
	return nil
}
