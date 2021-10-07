package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// ResponseType is a enum like type for response types
type ResponseType int

var responseTypes = [...]string{
	api.ResponseTypeToken,
	api.ResponseTypeShortToken,
	api.ResponseTypeTransferCode,
}

// ResponseTypes
const (
	ResponseTypeToken ResponseType = iota
	ResponseTypeShortToken
	ResponseTypeTransferCode
	maxResponseType
)

// NewResponseType creates a new ResponseType from the given response type string
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
		return errors.New("empty value in unmarshal response type")
	}
	*r = NewResponseType(s)
	if !r.Valid() {
		return errors.Errorf("value '%s' not valid for ResponseType", s)
	}
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (r *ResponseType) UnmarshalJSON(data []byte) error {
	var s string
	if err := errors.WithStack(json.Unmarshal(data, &s)); err != nil {
		return err
	}
	*r = NewResponseType(s)
	if !r.Valid() {
		return errors.Errorf("value '%s' not valid for ResponseType", s)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (r ResponseType) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(r.String())
	return data, errors.WithStack(err)
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

// Value implements the driver.Valuer interface.
func (r ResponseType) Value() (driver.Value, error) {
	return r.String(), nil
}

// Scan implements the sql.Scanner interface.
func (r *ResponseType) Scan(src interface{}) error {
	ns := sql.NullString{}
	if err := errors.WithStack(ns.Scan(src)); err != nil {
		return err
	}
	*r = NewResponseType(ns.String)
	return nil
}
