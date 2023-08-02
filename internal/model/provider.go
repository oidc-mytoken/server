package model

import (
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/pkg/oauth2x"
)

type Provider interface {
	Name() string
	Issuer() string
	ClientID() string
	Scopes() []string
	Endpoints() *oauth2x.Endpoints
	Audience() *AudienceConf
	MaxMytokenLifetime() int64
	AddClientAuthentication(r *resty.Request, endpoint string) *resty.Request
	GetAuthorizationURL(
		rlog log.Ext1FieldLogger, state, pkceChallenge string, scopeRestrictions, audRestrictions []string,
	) (string, error)
}

type AudienceConf struct {
	RFC8707           bool   `yaml:"use_rfc8707"`
	RequestParameter  string `yaml:"request_parameter"`
	SpaceSeparateAuds bool   `yaml:"space_separate_auds"`
}

const (
	AudienceParameterAudience = "audience"
	AudienceParameterResource = "resource"
)
