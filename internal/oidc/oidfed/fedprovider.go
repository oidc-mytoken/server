package oidfed

import (
	"net/url"

	oidfed "github.com/go-oidfed/lib"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/pkg/oauth2x"
)

func fedLeafEntity() *oidfed.FederationLeaf {
	return config.Get().Features.Federation.Entity
}

var defaultOIDFedAudienceConf = &model.AudienceConf{
	RFC8707:           true,
	RequestParameter:  model.AudienceParameterResource,
	SpaceSeparateAuds: false,
}

// OIDFedProvider implements the model.Provider interface for oidc fed
type OIDFedProvider struct {
	*oidfed.OpenIDProviderMetadata
}

// Name implements the model.Provider interface
func (p OIDFedProvider) Name() string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	if p.OrganizationName != "" {
		return p.OrganizationName
	}
	return p.OpenIDProviderMetadata.Issuer
}

// Issuer implements the model.Provider interface
func (p OIDFedProvider) Issuer() string {
	return p.OpenIDProviderMetadata.Issuer
}

// ClientID implements the model.Provider interface
func (OIDFedProvider) ClientID() string {
	return fedLeafEntity().EntityID()
}

// Scopes implements the model.Provider interface
func (p OIDFedProvider) Scopes() []string {
	return p.ScopesSupported
}

// Endpoints implements the model.Provider interface
func (p OIDFedProvider) Endpoints() *oauth2x.Endpoints {
	return &oauth2x.Endpoints{
		Authorization: p.AuthorizationEndpoint,
		Token:         p.TokenEndpoint,
		Userinfo:      p.UserinfoEndpoint,
		Registration:  p.RegistrationEndpoint,
		Revocation:    p.RevocationEndpoint,
		Introspection: p.IntrospectionEndpoint,
	}
}

// Audience implements the model.Provider interface
func (OIDFedProvider) Audience() *model.AudienceConf {
	return defaultOIDFedAudienceConf
}

// AccessTokenCache implements the model.Provider interface; it returns the default configuration for federated
// providers
func (OIDFedProvider) AccessTokenCache() *model.AccessTokenCacheConf {
	return config.Get().Features.Federation.AccessTokenCache
}

// MaxMytokenLifetime implements the model.Provider interface
func (OIDFedProvider) MaxMytokenLifetime() int64 {
	return 0
}

// AddClientAuthentication implements the model.Provider interface; it adds a client assertion to the request
func (OIDFedProvider) AddClientAuthentication(r *resty.Request, endpoint string) *resty.Request {
	clientAssertion, err := fedLeafEntity().RequestObjectProducer().ClientAssertion(endpoint)
	if err != nil {
		log.WithError(err).Error()
		return r
	}
	params := url.Values{}
	params.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	params.Set("client_assertion", string(clientAssertion))
	return r.SetFormDataFromValues(params)
}

// GetOIDFedProvider returns a OIDFedProvider implementing model.Provider for the passed issuer url
func GetOIDFedProvider(issuer string) model.Provider {
	meta, err := getOPMetadata(issuer)
	if err != nil {
		return nil
	}
	return OIDFedProvider{OpenIDProviderMetadata: meta}
}
