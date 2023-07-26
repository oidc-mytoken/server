package oidcfed

import (
	fed "github.com/zachmann/go-oidcfed/pkg"
)

func GetOPMetadata(issuer string) (*fed.OpenIDProviderMetadata, error) {
	return fedLeafEntity().ResolveOPMetadata(issuer)
}
