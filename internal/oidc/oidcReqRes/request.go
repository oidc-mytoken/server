package oidcReqRes

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/shared/utils"
)

// RefreshRequest is the oidc request for an refresh flow
type RefreshRequest struct {
	GrantType         string `json:"grant_type"`
	RefreshToken      string `json:"refresh_token"`
	Scopes            string `json:"scope,omitempty"`
	Audiences         string `json:"resource,omitempty"` // The "resource" key will be replaced with the string in resourceParameter
	resourceParameter string
}

// MarshalJSON implements the json.Marshaler interface
func (r *RefreshRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.ToFormData())
}

// NewRefreshRequest creates a new RefreshRequest for a given refresh token
func NewRefreshRequest(rt string, conf *config.ProviderConf) *RefreshRequest {
	return &RefreshRequest{
		GrantType:         "refresh_token",
		RefreshToken:      rt,
		resourceParameter: conf.AudienceRequestParameter,
	}
}

func parseTagValue(tag string) (value string, omitEmpty bool) {
	tmp := strings.Split(tag, ",")
	value = tmp[0]
	if len(tmp) > 1 {
		rest := tmp[1:]
		if utils.StringInSlice("omitempty", rest) {
			omitEmpty = true
		}
	}
	return
}

// ToFormData formats the RefreshRequest as a string map
func (r *RefreshRequest) ToFormData() map[string]string {
	v := reflect.ValueOf(*r)
	t := v.Type()
	m := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if k := f.Tag.Get("json"); k != "" {
			key, omitempty := parseTagValue(k)
			if key == "resource" {
				key = r.resourceParameter
			}
			value := v.Field(i).String()
			if !omitempty || value != "" {
				m[key] = value
			}
		}
	}
	return m
}

// RevokeRequest is a oidc request for revoking tokens
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
	return utils.StructToStringMapUsingJSONTags(r)
}
