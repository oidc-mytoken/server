package oidfed

import (
	oidfed "github.com/go-oidfed/lib"
)

// getOPMetadata returns the fed.OpenIDProviderMetadata for an oidfed issuer
func getOPMetadata(issuer string) (*oidfed.OpenIDProviderMetadata, error) {
	return fedLeafEntity().ResolveOPMetadata(issuer)
}
