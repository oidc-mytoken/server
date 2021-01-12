package model

import (
	"strings"

	"github.com/zachmann/mytoken/internal/server/utils/issuerUtils"
)

type Provider struct {
	Name         string `yaml:"name"`
	Issuer       string `yaml:"url"`
	GPGKey       string `yaml:"default_gpg_key"`
	DefaultToken string `yaml:"default_token"`
}

func NewProviderFromString(provider string) *Provider {
	p := &Provider{}
	if strings.HasPrefix(provider, "https://") {
		p.Issuer = provider
	} else {
		p.Name = provider
	}
	return p
}

type Providers []*Provider

func (p Provider) Equals(b Provider, compUrl bool) bool {
	if compUrl {
		return p.Compare(b.Issuer, compUrl)
	}
	return p.Compare(b.Name, compUrl)
}
func (p Provider) Compare(b string, compUrl bool) bool {
	if compUrl {
		return issuerUtils.CompareIssuerURLs(p.Issuer, b)
	}
	return p.Name == b
}

func (p Providers) Find(provider string) (*Provider, bool) {
	isURL := strings.HasPrefix(provider, "https://")
	return p.FindBy(provider, isURL)
}

func (p Providers) FindBy(provider string, compURL bool) (*Provider, bool) {
	key := NewProviderFromString(provider)
	for _, pp := range p {
		if pp.Equals(*key, compURL) {
			return pp, true
		}
	}
	return key, false
}
