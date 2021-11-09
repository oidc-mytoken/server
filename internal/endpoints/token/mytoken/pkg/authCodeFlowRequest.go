package pkg

import (
	"encoding/json"

	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
)

// AuthCodeFlowRequest holds a authorization code flow request
type AuthCodeFlowRequest struct {
	OIDCFlowRequest
	RedirectType string `json:"redirect_type"`
}

// Native checks if the request is native
func (r *AuthCodeFlowRequest) Native() bool {
	return r.RedirectType != api.RedirectTypeWeb
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
	r.redirectType = r.RedirectType
	data, err := json.Marshal(r.OIDCFlow)
	return data, errors.WithStack(err)
}
