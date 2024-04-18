package oidcreqres

import (
	"net/url"
	"strings"

	"github.com/oidc-mytoken/server/internal/model"
	iutils "github.com/oidc-mytoken/server/internal/utils"
)

// RefreshRequest is the oidc request for a refresh flow
type RefreshRequest struct {
	GrantType         string
	RefreshToken      string
	Scopes            string
	Audiences         []string
	resourceParameter string
	spaceDelimited    bool
}

// NewRefreshRequest creates a new RefreshRequest for a given refresh token
func NewRefreshRequest(rt string, aud *model.AudienceConf) *RefreshRequest {
	if aud == nil {
		aud = &model.AudienceConf{
			RFC8707:           true,
			RequestParameter:  model.AudienceParameterResource,
			SpaceSeparateAuds: false,
		}
	}
	return &RefreshRequest{
		GrantType:         "refresh_token",
		RefreshToken:      rt,
		resourceParameter: aud.RequestParameter,
		spaceDelimited:    aud.SpaceSeparateAuds,
	}
}

// ToURLValues formats the RefreshRequest as an url.Values
func (r *RefreshRequest) ToURLValues() url.Values {
	m := make(url.Values)
	m["grant_type"] = []string{r.GrantType}
	m["refresh_token"] = []string{r.RefreshToken}
	if r.Scopes != "" {
		m["scope"] = []string{r.Scopes}
	}
	if len(r.Audiences) > 0 && r.Audiences[0] != "" {
		if r.spaceDelimited {
			m[r.resourceParameter] = []string{strings.Join(r.Audiences, " ")}
		} else {
			m[r.resourceParameter] = r.Audiences
		}
	}
	return m
}

// RevokeRequest is an oidc request for revoking tokens
type RevokeRequest struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type_hint"`
}

// NewRTRevokeRequest creates a new RevokeRequest for revoking the passed refresh token
func NewRTRevokeRequest(rt string) *RevokeRequest {
	return &RevokeRequest{
		Token:     rt,
		TokenType: "refresh_token",
	}
}

// ToFormData formats the RevokeRequest as a string map
func (r *RevokeRequest) ToFormData() map[string]string {
	return iutils.StructToStringMapUsingJSONTags(r)
}
