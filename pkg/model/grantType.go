package model

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// GrantType is an enum like type for grant types
type GrantType int

// AllGrantTypes holds all defined GrantType strings
var AllGrantTypes = [...]string{"mytoken", "oidc_flow", "polling_code", "access_token", "private_key_jwt", "transfer_code"}

// GrantTypes
const (
	GrantTypeMytoken GrantType = iota
	GrantTypeOIDCFlow
	GrantTypePollingCode
	GrantTypeAccessToken
	GrantTypePrivateKeyJWT
	GrantTypeTransferCode
	maxGrantType
)

// NewGrantType creates a new GrantType from the grant type string
func NewGrantType(s string) GrantType {
	for i, f := range AllGrantTypes {
		if f == s {
			return GrantType(i)
		}
	}
	return -1
}

func (g *GrantType) String() string {
	if *g < 0 || int(*g) >= len(AllGrantTypes) {
		return ""
	}
	return AllGrantTypes[*g]
}

// Valid checks that GrantType is a defined grant type
func (g *GrantType) Valid() bool {
	return *g < maxGrantType && *g >= 0
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (g *GrantType) UnmarshalYAML(value *yaml.Node) error {
	s := value.Value
	if s == "" {
		return fmt.Errorf("empty value in unmarshal grant type")
	}
	*g = NewGrantType(s)
	if !g.Valid() {
		return fmt.Errorf("value '%s' not valid for GrantType", s)
	}
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (g *GrantType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*g = NewGrantType(s)
	if !g.Valid() {
		return fmt.Errorf("value '%s' not valid for GrantType", s)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (g GrantType) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.String())
}

// AddToSliceIfNotFound adds the GrantType to a slice s if it is not already there
func (g GrantType) AddToSliceIfNotFound(s *[]GrantType) {
	for _, ss := range *s {
		if ss == g {
			return
		}
	}
	*s = append(*s, g)
}
