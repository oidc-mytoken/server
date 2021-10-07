package model

import (
	"encoding/json"

	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// OIDCFlow is a enum like type for oidc flows
type OIDCFlow int

var oidcFlows = [...]string{api.OIDCFlowAuthorizationCode}

// OIDCFlows
const (
	OIDCFlowAuthorizationCode OIDCFlow = iota
	// OIDCFlowDevice
	maxFlow
)

// NewOIDCFlow creates a new OIDCFlow from the flow string
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
		return errors.New("empty value in unmarshal oidc flow")
	}
	*f = NewOIDCFlow(s)
	if !f.Valid() {
		return errors.Errorf("value '%s' not valid for OIDCFlow", s)
	}
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (f *OIDCFlow) UnmarshalJSON(data []byte) error {
	var s string
	if err := errors.WithStack(json.Unmarshal(data, &s)); err != nil {
		return err
	}
	*f = NewOIDCFlow(s)
	if !f.Valid() {
		return errors.Errorf("value '%s' not valid for OIDCFlow", s)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (f OIDCFlow) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(f.String())
	return data, errors.WithStack(err)
}

// AddToSliceIfNotFound adds the OIDCFlow to a slice s if it is not already there
func (f OIDCFlow) AddToSliceIfNotFound(s *[]OIDCFlow) {
	if OIDCFlowIsInSlice(f, *s) {
		return
	}
	*s = append(*s, f)
}

// OIDCFlowIsInSlice checks if a OIDCFlow is present in a slice of OIDCFlows
func OIDCFlowIsInSlice(f OIDCFlow, s []OIDCFlow) bool {
	for _, ss := range s {
		if ss == f {
			return true
		}
	}
	return false
}
