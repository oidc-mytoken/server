package oidcfed

import (
	"time"

	"github.com/pkg/errors"
	fed "github.com/zachmann/go-oidcfed/pkg"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/utils/cache"
)

func GetOPMetadata(issuer string) (*fed.OpenIDProviderMetadata, error) {
	v, set := cache.Get(cache.FederationOPMetadata, issuer)
	if set {
		opm, ok := v.(*fed.OpenIDProviderMetadata)
		if ok {
			return opm, nil
		}
	}
	tr := fed.TrustResolver{
		TrustAnchors:   config.Get().Features.Federation.TrustAnchors,
		StartingEntity: issuer,
	}
	chains := tr.ResolveToValidChains()
	chains = chains.Filter(fed.TrustChainsFilterMinPathLength)
	if len(chains) == 0 {
		return nil, errors.New("no trust chain found")
	}
	chain := chains[0]
	m, err := chain.Metadata()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	delta := time.Unix(chain.ExpiresAt(), 0).Sub(time.Now()) - time.Minute // we subtract a one-minute puffer
	if delta > 0 {
		cache.Set(cache.FederationOPMetadata, issuer, m.OpenIDProvider, delta)
	}
	return m.OpenIDProvider, nil
}
