package oidcfed

import (
	"fmt"
	"time"

	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"
	oidcfed "github.com/zachmann/go-oidcfed/pkg"

	"github.com/oidc-mytoken/server/internal/config"
)

var discoverer = oidcfed.FilterableVerifiedChainsOPDiscoverer{
	Filters: []oidcfed.OPDiscoveryFilter{
		oidcfed.OPDiscoveryFilterSupportedGrantTypesIncludes("refresh_token"),
		oidcfed.OPDiscoveryFilterSupportedScopesIncludes("offline_access"),
	},
}

var oidcfedIssuers []string

var ticker *time.Ticker

// Discovery starts the OP discovery process for OPs below the configured trust anchors and also schedules a rerun
func Discovery() {
	if !config.Get().Features.Federation.Enabled {
		return
	}
	discovery()
	if ticker != nil {
		ticker.Reset(time.Hour)
		return
	}
	ticker = time.NewTicker(time.Hour)
	go func() {
		for range ticker.C {
			discovery()
		}
	}()
}

func discovery() {
	log.Debug("Running oidcfed OP discovery")
	opInfos := discoverer.Discover(config.Get().Features.Federation.TrustAnchors...)
	tmp := make([]string, len(opInfos))
	for i, op := range opInfos {
		tmp[i] = op.Issuer
	}
	oidcfedIssuers = tmp
}

// Issuers returns a slice of issuer urls of OPs discovered in the federation
func Issuers() []string {
	return oidcfedIssuers
}

// SupportedProviders return the api.SupportedProviderConfig for the discovered OPs in the federation
func SupportedProviders() (providers []api.SupportedProviderConfig) {
	names := make(map[string][]int)
	for index, issuer := range oidcfedIssuers {
		p := GetOIDCFedProvider(issuer)
		if p == nil {
			log.WithField("issuer", issuer).Error("error while obtaining op metadata in federation")
			continue
		}
		names[p.Name()] = append(names[p.Name()], index)
		providers = append(
			providers, api.SupportedProviderConfig{
				Issuer:          p.Issuer(),
				Name:            p.Name(),
				ScopesSupported: p.Scopes(),
				OIDCFed:         true,
			},
		)
	}
	for _, indices := range names {
		if len(indices) <= 1 {
			continue
		}
		for _, i := range indices {
			p := providers[i]
			p.Name = fmt.Sprintf("%s (%s)", p.Name, p.Issuer)
			providers[i] = p
		}
	}
	return
}
