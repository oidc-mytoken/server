package oidcfed

import (
	"net/url"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	fed "github.com/zachmann/go-oidfed/pkg"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/pkg/oauth2x"
)

func fedLeafEntity() *fed.FederationLeaf {
	return config.Get().Features.Federation.Entity
}

var defaultOIDCFedAudienceConf = &model.AudienceConf{
	RFC8707:           true,
	RequestParameter:  model.AudienceParameterResource,
	SpaceSeparateAuds: false,
}

// OIDCFedProvider implements the model.Provider interface for oidc fed
type OIDCFedProvider struct {
	*fed.OpenIDProviderMetadata
}

// Name implements the model.Provider interface
func (p OIDCFedProvider) Name() string {
	if p.OrganizationName != "" {
		return p.OrganizationName
	}
	return p.OpenIDProviderMetadata.Issuer
}

// Issuer implements the model.Provider interface
func (p OIDCFedProvider) Issuer() string {
	return p.OpenIDProviderMetadata.Issuer
}

// ClientID implements the model.Provider interface
func (OIDCFedProvider) ClientID() string {
	return fedLeafEntity().EntityID
}

// Scopes implements the model.Provider interface
func (p OIDCFedProvider) Scopes() []string {
	return p.ScopesSupported
}

// Endpoints implements the model.Provider interface
func (p OIDCFedProvider) Endpoints() *oauth2x.Endpoints {
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
func (OIDCFedProvider) Audience() *model.AudienceConf {
	return defaultOIDCFedAudienceConf
}

// MaxMytokenLifetime implements the model.Provider interface
func (OIDCFedProvider) MaxMytokenLifetime() int64 {
	return 0
}

// AddClientAuthentication implements the model.Provider interface; it adds a client assertion to the request
func (OIDCFedProvider) AddClientAuthentication(r *resty.Request, endpoint string) *resty.Request {
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

// GetOIDCFedProvider returns a OIDCFedProvider implementing model.Provider for the passed issuer url
func GetOIDCFedProvider(issuer string) model.Provider {
	meta, err := getOPMetadata(issuer)
	if err != nil {
		return nil
	}
	return OIDCFedProvider{meta}
}
