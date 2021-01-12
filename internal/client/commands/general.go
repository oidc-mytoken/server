package commands

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/client/config"
	"github.com/zachmann/mytoken/internal/client/model"
)

// generalOptions holds command line options that can be used with all commands
type generalOptions struct {
	Provider   string `short:"p" long:"provider" description:"The name or issuer url of the OpenID provider that should be used"`
	Name       string `short:"t" long:"name" description:"The name of the super token that should be used"`
	SuperToken string `long:"ST" description:"The passed super token is used instead of a stored one"`
}

type providerOpt string

func (g *generalOptions) Check() (*model.Provider, string) {
	p, pErr := g.checkProvider()
	if len(g.SuperToken) > 0 {
		return p, g.SuperToken
	}
	if pErr != nil {
		log.Fatal(pErr)
	}
	token, tErr := config.Get().GetToken(p.Issuer, g.Name)
	if tErr != nil {
		log.Fatal(tErr)
	}
	return p, token
}

func (g *generalOptions) checkToken(issuer string) (string, error) {
	if len(g.SuperToken) > 0 {
		return g.SuperToken, nil
	}
	return config.Get().GetToken(issuer, g.Name)
}

func (g *generalOptions) checkProvider() (p *model.Provider, err error) {
	provider := g.Provider
	if len(provider) == 0 {
		provider = config.Get().DefaultProvider
		if len(provider) == 0 {
			err = fmt.Errorf("Provider not specified and no default provider set")
			return
		}
	}
	isURL := strings.HasPrefix(provider, "https://")
	pp, ok := config.Get().Providers.FindBy(provider, isURL)
	if !ok && !isURL {
		err = fmt.Errorf("Provider name '%s' not found in config file. Please provide a valid provider name or the provider url.", provider)
		return
	}
	return pp, nil
}
