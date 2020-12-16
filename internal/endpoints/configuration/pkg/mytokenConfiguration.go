package pkg

import "github.com/zachmann/mytoken/internal/model"

// MytokenConfiguration holds information about a mytoken instance
type MytokenConfiguration struct {
	Issuer                                 string                    `json:"issuer"`
	AccessTokenEndpoint                    string                    `json:"access_token_endpoint"`
	SuperTokenEndpoint                     string                    `json:"super_token_endpoint"`
	TokeninfoEndpoint                      string                    `json:"tokeninfo_endpoint,omitempty"`
	RevocationEndpoint                     string                    `json:"revocation_endpoint,omitempty"`
	UserSettingsEndpoint                   string                    `json:"usersettings_endpoint"`
	TokenTransferEndpoint                  string                    `json:"token_transfer_endpoint,omitempty"`
	JWKSURI                                string                    `json:"jwks_uri"`
	ProvidersSupported                     []SupportedProviderConfig `json:"providers_supported"`
	TokenSigningAlgValue                   string                    `json:"token_signing_alg_value"`
	AccessTokenEndpointGrantTypesSupported []model.GrantType         `json:"access_token_endpoint_grant_types_supported"`
	SuperTokenEndpointGrantTypesSupported  []model.GrantType         `json:"super_token_endpoint_grant_types_supported"`
	SuperTokenEndpointOIDCFlowsSupported   []model.OIDCFlow          `json:"super_token_endpoint_oidc_flow_supported"`
	ResponseTypesSupported                 []model.ResponseType      `json:"response_types_supported"`
	ServiceDocumentation                   string                    `json:"service_documentation,omitempty"`
	Version                                string                    `json:"version,omitempty"`
}

// SupportedProviderConfig holds information about a provider
type SupportedProviderConfig struct {
	Issuer          string   `json:"issuer"`
	ScopesSupported []string `json:"scopes_supported"`
}
