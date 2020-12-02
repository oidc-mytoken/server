package model

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

type OIDCFlow int

var oidcFlows = [...]string{"authorization_code", "device"}

// OIDCFlows
const (
	OIDCFlowAuthorizationCode OIDCFlow = iota
	OIDCFlowDevice
	maxFlow
)

func NewOIDCFlow(s string) OIDCFlow {
	for i, f := range oidcFlows {
		if f == s {
			return OIDCFlow(i)
		}
	}
	return -1
}

func (f *OIDCFlow) String() string {
	if *f < 0 || int(*f) >= len(oidcFlows) {
		return ""
	}
	return oidcFlows[*f]
}

// Valid checks that OIDCFlow is a defined flow
func (f *OIDCFlow) Valid() bool {
	return *f < maxFlow && *f >= 0
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (f *OIDCFlow) UnmarshalYAML(value *yaml.Node) error {
	s := value.Value
	if s == "" {
		return fmt.Errorf("empty value in unmarshal oidc flow")
	}
	*f = NewOIDCFlow(s)
	if !f.Valid() {
		return fmt.Errorf("value '%s' not valid for OIDCFlow", s)
	}
	return nil
}

func (f *OIDCFlow) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*f = NewOIDCFlow(s)
	if !f.Valid() {
		return fmt.Errorf("value '%s' not valid for OIDCFlow", s)
	}
	return nil
}

func (f *OIDCFlow) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

// AddToSliceIfNotFound adds the OIDCFlow to a slice s if it is not already there
func (f OIDCFlow) AddToSliceIfNotFound(s []OIDCFlow) {
	if OIDCFlowIsInSlice(f, s) {
		return
	}
	s = append(s, f)
}

func OIDCFlowIsInSlice(f OIDCFlow, s []OIDCFlow) bool {
	for _, ss := range s {
		if ss == f {
			return true
		}
	}
	return false
}
