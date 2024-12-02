package pkg

import (
	"database/sql/driver"
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
)

// AuthCodeFlowRequest holds a authorization code flow request
type AuthCodeFlowRequest struct {
	OIDCFlowRequest
	AuthCodeFlowAttrs
}

// AuthCodeFlowAttrs holds the additional attributes for AuthCodeFlowRequests
type AuthCodeFlowAttrs struct {
	ClientType  string `json:"client_type"`
	RedirectURI string `json:"redirect_uri"`
}

// Native checks if the request is native
func (r AuthCodeFlowRequest) Native() bool {
	return r.ClientType != api.ClientTypeWeb
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (r *AuthCodeFlowRequest) UnmarshalJSON(data []byte) error {
	if err := errors.WithStack(json.Unmarshal(data, &r.OIDCFlowRequest)); err != nil {
		return err
	}
	return errors.WithStack(json.Unmarshal(data, &r.AuthCodeFlowAttrs))
}

// MarshalJSON implements the json.Marshaler interface
func (r AuthCodeFlowRequest) MarshalJSON() ([]byte, error) {
	base, err := json.Marshal(r.OIDCFlowRequest)
	if err != nil {
		return nil, err
	}
	attrs, err := json.Marshal(r.AuthCodeFlowAttrs)
	if err != nil {
		return nil, err
	}
	return jsonpatch.MergePatch(base, attrs)
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
func (r AuthCodeFlowRequest) Value() (driver.Value, error) { // skipcq: CRT-P0003, RVV-B0006
	r.IncludedProfiles = nil
	v, err := json.Marshal(r)
	return v, errors.WithStack(err)
}
