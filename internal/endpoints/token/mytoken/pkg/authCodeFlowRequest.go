package pkg

import (
	"encoding/json"
)

// Redirect types
const (
	redirectTypeWeb    = "web"
	redirectTypeNative = "native"
)

// AuthCodeFlowRequest holds a authorization code flow request
type AuthCodeFlowRequest struct {
	OIDCFlowRequest
	RedirectType string `json:"redirect_type"`
}

// Native checks if the request is native
func (r *AuthCodeFlowRequest) Native() bool {
	return r.RedirectType == redirectTypeNative
}

// UnmarshalJSON implements the json unmarshaler interface
func (r *AuthCodeFlowRequest) UnmarshalJSON(data []byte) error {
	var tmp OIDCFlowRequest
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*r = tmp.ToAuthCodeFlowRequest()
	return nil
}

// MarshalJSON implements the json marshaler interface
func (r *AuthCodeFlowRequest) MarshalJSON() ([]byte, error) {
	r.redirectType = r.RedirectType
	return json.Marshal(r.OIDCFlow)
}
