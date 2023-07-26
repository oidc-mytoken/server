package oidcfed

import (
	"net/url"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	fed "github.com/zachmann/go-oidcfed/pkg"

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

type OIDCFedProvider struct {
	*fed.OpenIDProviderMetadata
}

func (p OIDCFedProvider) Name() string {
	if p.OrganizationName != "" {
		return p.OrganizationName
	}
	return p.OpenIDProviderMetadata.Issuer
}

func (p OIDCFedProvider) Issuer() string {
	return p.OpenIDProviderMetadata.Issuer
}

func (p OIDCFedProvider) ClientID() string {
	return fedLeafEntity().EntityID
}

func (p OIDCFedProvider) Scopes() []string {
	return p.ScopesSupported
}

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

func (p OIDCFedProvider) Audience() *model.AudienceConf {
	return defaultOIDCFedAudienceConf
}

func (p OIDCFedProvider) MaxMytokenLifetime() int64 {
	return 0
}

func (p OIDCFedProvider) AddClientAuthentication(r *resty.Request, endpoint string) *resty.Request {
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

func GetOIDCFedProvider(issuer string) model.Provider {
	meta, err := GetOPMetadata(issuer)
	if err != nil {
		return nil
	}
	return OIDCFedProvider{meta}
}
