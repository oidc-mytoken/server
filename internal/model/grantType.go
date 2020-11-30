package model

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type GrantType int

var grantTypes = [...]string{"super_token", "oidc_flow", "polling_code", "access_token", "private_key_jwt"}

// GrantTypes
const (
	GrantTypeSuperToken GrantType = iota
	GrantTypeOIDCFlow
	GrantTypePollingCode
	GrantTypeAccessToken
	GrantTypePrivateKeyJWT
	maxGrantType
)

func NewGrantType(s string) GrantType {
	log.WithField("grant_type", s).Trace("Grant Type")
	for i, f := range grantTypes {
		if f == s {
			return GrantType(i)
		}
	}
	return -1
}

func (g *GrantType) String() string {
	if *g < 0 || int(*g) >= len(grantTypes) {
		return ""
	}
	return grantTypes[*g]
}

// Valid checks that GrantType is a defined flow
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

func (g *GrantType) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.String())
}

// AddToSliceIfNotFound adds the GrantType to a slice s if it is not already there
func (g GrantType) AddToSliceIfNotFound(s []GrantType) {
	for _, ss := range s {
		if ss == g {
			return
		}
	}
	s = append(s, g)
}
