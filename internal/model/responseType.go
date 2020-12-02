package model

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

type ResponseType int

var responseTypes = [...]string{"token", "short_token", "transfer_code"}

// ResponseTypes
const (
	ResponseTypeToken ResponseType = iota
	ResponseTypeShortToken
	ResponseTypeTransferCode
	maxResponseType
)

func NewResponseType(s string) ResponseType {
	for i, f := range responseTypes {
		if f == s {
			return ResponseType(i)
		}
	}
	return -1
}

func (r *ResponseType) String() string {
	if *r < 0 || int(*r) >= len(responseTypes) {
		return ""
	}
	return responseTypes[*r]
}

// Valid checks that ResponseType is a defined flow
func (r *ResponseType) Valid() bool {
	return *r < maxResponseType && *r >= 0
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (r *ResponseType) UnmarshalYAML(value *yaml.Node) error {
	s := value.Value
	if s == "" {
		return fmt.Errorf("empty value in unmarshal response type")
	}
	*r = NewResponseType(s)
	if !r.Valid() {
		return fmt.Errorf("value '%s' not valid for ResponseType", s)
	}
	return nil
}

func (r *ResponseType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*r = NewResponseType(s)
	if !r.Valid() {
		return fmt.Errorf("value '%s' not valid for ResponseType", s)
	}
	return nil
}

func (r ResponseType) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// AddToSliceIfNotFound adds the ResponseType to a slice s if it is not already there
func (r ResponseType) AddToSliceIfNotFound(s *[]ResponseType) {
	for _, ss := range *s {
		if ss == r {
			return
		}
	}
	*s = append(*s, r)
}
