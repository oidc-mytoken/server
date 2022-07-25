package pkg

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
)

// AuthCodeFlowRequest holds a authorization code flow request
type AuthCodeFlowRequest struct {
	OIDCFlowRequest
	ClientType  string `json:"client_type"`
	RedirectURI string `json:"redirect_uri"`
}

// Native checks if the request is native
func (r *AuthCodeFlowRequest) Native() bool {
	return r.ClientType != api.ClientTypeWeb
}

// UnmarshalJSON implements the json unmarshaler interface
func (r *AuthCodeFlowRequest) UnmarshalJSON(data []byte) error {
	var tmp OIDCFlowRequest
	if err := json.Unmarshal(data, &tmp); err != nil {
		return errors.WithStack(err)
	}
	*r = tmp.ToAuthCodeFlowRequest()
	return nil
}

// MarshalJSON implements the json marshaler interface
func (r *AuthCodeFlowRequest) MarshalJSON() ([]byte, error) {
	r.clientType = r.ClientType
	r.redirectURI = r.RedirectURI
	data, err := json.Marshal(r.OIDCFlow)
	return data, errors.WithStack(err)
}

// Scan implements the sql.Scanner interface
func (r *AuthCodeFlowRequest) Scan(src interface{}) error {
	v, ok := src.([]byte)
	if !ok {
		return errors.New("bad []byte type assertion")
	}
	return errors.WithStack(json.Unmarshal(v, r))
}

// Value implements the driver.Valuer interface
func (r AuthCodeFlowRequest) Value() (driver.Value, error) { // skipcq: CRT-P0003
	v, err := json.Marshal(r)
	return v, errors.WithStack(err)
}
