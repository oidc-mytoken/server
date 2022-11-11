package oauth2x

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

// Config is the configuration for an oauth2 provider, especially all relevant endpoint
type Config struct {
	Issuer    string
	Ctx       context.Context
	endpoints *Endpoints
}

// Endpoints holds all relevant OAuth2/OIDC endpoints
type Endpoints struct {
	Authorization string `json:"authorization_endpoint"`
	Token         string `json:"token_endpoint"`
	Userinfo      string `json:"userinfo_endpoint"`
	Registration  string `json:"registration_endpoint"`
	Revocation    string `json:"revocation_endpoint"`
	Introspection string `json:"introspection_endpoint"`
}

// OAuth2 returns the endpoints as oauth2.Endpoint so it can be used with the oauth2 package
func (e *Endpoints) OAuth2() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  e.Authorization,
		TokenURL: e.Token,
	}
}

// Endpoints returns the Endpoints for this Config
func (c *Config) Endpoints() (*Endpoints, error) {
	var err error
	if c.endpoints == nil {
		err = c.discovery()
	}
	return c.endpoints, err
}

// NewConfig creates a new Config for the passed issuer with the passed context.Context and performs the OAuth2/OpenID
// discovery
func NewConfig(ctx context.Context, issuer string) (*Config, error) {
	c := &Config{
		Issuer: issuer,
		Ctx:    ctx,
	}
	err := c.discovery()
	return c, err
}

func (c *Config) discovery() error {
	wellKnown := strings.TrimSuffix(c.Issuer, "/") + "/.well-known/openid-configuration"
	req, err := http.NewRequest("GET", wellKnown, nil)
	if err != nil {
		return err
	}
	resp, err := doRequest(c.Ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", resp.Status, body)
	}

	var endpoints Endpoints
	if err = json.Unmarshal(body, &endpoints); err != nil {
		return fmt.Errorf("oauth2x: failed to decode provider discovery object: %v", err)
	}
	c.endpoints = &endpoints
	return nil
}

func doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	client := http.DefaultClient
	if c, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok {
		client = c
	}
	return client.Do(req.WithContext(ctx))
}
