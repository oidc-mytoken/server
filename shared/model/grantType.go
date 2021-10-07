package model

import (
	"encoding/json"

	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// GrantType is an enum like type for grant types
type GrantType int

// AllGrantTypes holds all defined GrantType strings
var AllGrantTypes = api.AllGrantTypes

// GrantTypes
const ( // assert that these are in the same order as api.AllGrantTypes
	GrantTypeMytoken GrantType = iota
	GrantTypeOIDCFlow
	GrantTypePollingCode
	GrantTypeTransferCode
	GrantTypeSSH
	maxGrantType
)

// NewGrantType creates a new GrantType from the grant type string
func NewGrantType(s string) GrantType {
	for i, f := range AllGrantTypes {
		if f == s {
			return GrantType(i)
		}
	}
	if s == "refresh_token" { // RT=MT compatibility
		return GrantTypeMytoken
	}
	return -1
}

const valueNotValidFmt = "value '%s' not valid for GrantType"

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
		return errors.New("empty value in unmarshal grant type")
	}
	*g = NewGrantType(s)
	if !g.Valid() {
		return errors.Errorf(valueNotValidFmt, s)
	}
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (g *GrantType) UnmarshalJSON(data []byte) error {
	var s string
	if err := errors.WithStack(json.Unmarshal(data, &s)); err != nil {
		return err
	}
	*g = NewGrantType(s)
	if !g.Valid() {
		return errors.Errorf(valueNotValidFmt, s)
	}
	return nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (g *GrantType) UnmarshalText(data []byte) error {
	s := string(data)
	*g = NewGrantType(s)
	if !g.Valid() {
		return errors.Errorf(valueNotValidFmt, s)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (g GrantType) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(g.String())
	return data, errors.WithStack(err)
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
