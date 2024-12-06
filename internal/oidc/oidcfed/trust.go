package oidcfed

import (
	fed "github.com/zachmann/go-oidfed/pkg"
)

// getOPMetadata returns the fed.OpenIDProviderMetadata for an oidcfed issuer
func getOPMetadata(issuer string) (*fed.OpenIDProviderMetadata, error) {
	return fedLeafEntity().ResolveOPMetadata(issuer)
}
