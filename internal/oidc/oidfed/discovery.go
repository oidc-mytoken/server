package oidfed

import (
	"fmt"
	"time"

	oidfed "github.com/go-oidfed/lib"
	"github.com/go-oidfed/lib/apimodel"
	"github.com/go-oidfed/lib/oidfedconst"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
)

var oidfedOPs map[string]*oidfed.CollectedEntity
var ticker *time.Ticker

// Discovery starts the OP discovery process for OPs below the configured trust anchors
// and schedules periodic reruns based on the configured interval
func Discovery() {
	if !config.Get().Features.Federation.Enabled {
		return
	}

	discovery()

	interval := time.Duration(config.Get().Features.Federation.OPDiscovery.Interval) * time.Second

	if ticker != nil {
		ticker.Reset(interval)
		return
	}
	ticker = time.NewTicker(interval)
	go func() {
		for range ticker.C {
			discovery()
		}
	}()
}

func discovery() {
	log.Debug("Running oidfed OP discovery")

	trustAnchors := config.Get().Features.Federation.TrustAnchors
	trustAnchorIDs := trustAnchors.EntityIDs()
	opDiscoveryConf := config.Get().Features.Federation.OPDiscovery

	// Build filters
	filters := []oidfed.EntityCollectionFilter{
		oidfed.EntityCollectionFilterOPSupportedGrantTypesIncludes(trustAnchorIDs, "refresh_token"),
		oidfed.EntityCollectionFilterOPSupportedScopesIncludes(trustAnchorIDs, "offline_access"),
		oidfed.EntityCollectionFilterOPSupportsAutomaticRegistration(trustAnchorIDs),
	}

	// Add trust mark filter if required trust marks are configured
	if len(opDiscoveryConf.RequiredTrustMarks) > 0 {
		filters = append(
			filters, oidfed.NewEntityCollectionFilter(
				func(e *oidfed.CollectedEntity) bool {
					ok, err := oidfed.VerifyEntityHasValidTrustmarks(
						e.EntityID,
						opDiscoveryConf.RequiredTrustMarks,
						trustAnchors,
					)
					if err != nil {
						log.WithError(err).WithField("entity_id", e.EntityID).
							Error("error during trustmark verification")
					}
					return ok
				},
			),
		)
	}

	// Choose collector based on config
	var collector oidfed.EntityCollector
	if opDiscoveryConf.UseEntityCollectionEndpoint {
		collector = oidfed.SmartRemoteEntityCollector{TrustAnchors: trustAnchorIDs}
	} else {
		collector = &oidfed.SimpleEntityCollector{}
	}

	// Collect OPs from all trust anchors
	providers := make(map[string]*oidfed.CollectedEntity)
	for _, ta := range trustAnchors {
		response, errResp := oidfed.FilterableVerifiedChainsEntityCollector{
			Collector: collector,
			Filters:   filters,
		}.CollectEntities(
			apimodel.EntityCollectionRequest{
				TrustAnchor: ta.EntityID,
				EntityTypes: []string{oidfedconst.EntityTypeOpenIDProvider},
			},
		)

		if errResp != nil {
			log.WithField("trust_anchor", ta.EntityID).
				WithField("error", errResp.Error).
				Error("error collecting OPs from trust anchor")
			continue
		}

		if response != nil {
			for _, op := range response.FederationEntities {
				providers[op.EntityID] = op
			}
		}
	}

	oidfedOPs = providers

	log.WithField("count", len(providers)).Debug("OP discovery completed")
}

func getDisplayNameFromEntityInfo(entity *oidfed.CollectedEntity) string {
	if entity == nil {
		return ""
	}
	if entity.UIInfos == nil {
		return entity.EntityID
	}
	op, ok := entity.UIInfos[oidfedconst.EntityTypeOpenIDProvider]
	if ok && op.DisplayName != "" {
		return op.DisplayName
	}
	fed, ok := entity.UIInfos[oidfedconst.EntityTypeFederationEntity]
	if ok && fed.DisplayName != "" {
		return fed.DisplayName
	}
	return entity.EntityID
}

// SupportedProviders returns the api.SupportedProviderConfig for the discovered OPs in the federation
func SupportedProviders() (providers []api.SupportedProviderConfig) {
	names := make(map[string][]int)
	i := 0
	for issuer, entity := range oidfedOPs {
		p := GetOIDFedProvider(issuer)
		if p == nil {
			log.WithField("issuer", issuer).Error("error while obtaining op metadata in federation")
			continue
		}
		name := getDisplayNameFromEntityInfo(entity)
		names[p.Name()] = append(names[p.Name()], i)
		providers = append(
			providers, api.SupportedProviderConfig{
				Issuer:          entity.EntityID,
				Name:            name,
				ScopesSupported: p.Scopes(),
				OIDFed:          true,
			},
		)
		i++
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
