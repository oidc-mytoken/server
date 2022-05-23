package pkg

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
)

// OIDCFlowRequest holds the request for an OIDC Flow request
type OIDCFlowRequest struct {
	api.OIDCFlowRequest `json:",inline"`
	GrantType           model.GrantType           `json:"grant_type"`
	OIDCFlow            model.OIDCFlow            `json:"oidc_flow"`
	Restrictions        restrictions.Restrictions `json:"restrictions"`
	ResponseType        model.ResponseType        `json:"response_type"`
	clientType          string
	redirectURI         string
}

// NewOIDCFlowRequest creates a new OIDCFlowRequest with default values where they can be omitted
func NewOIDCFlowRequest() *OIDCFlowRequest {
	return &OIDCFlowRequest{
		OIDCFlowRequest: api.OIDCFlowRequest{
			GeneralMytokenRequest: api.GeneralMytokenRequest{
				Capabilities: api.Capabilities{api.CapabilityAT},
			},
		},
		ResponseType: model.ResponseTypeToken,
		clientType:   api.ClientTypeNative,
		GrantType:    -1,
	}
}

// SetRedirectType sets the (hidden) redirect type
func (r *OIDCFlowRequest) SetRedirectType(redirect string) {
	r.clientType = redirect
}

// SetRedirectURI sets the (hidden) redirect uri
func (r *OIDCFlowRequest) SetRedirectURI(uri string) {
	r.redirectURI = uri
}

// MarshalJSON implements the json.Marshaler interface
func (r OIDCFlowRequest) MarshalJSON() ([]byte, error) {
	type ofr OIDCFlowRequest
	o := struct {
		ofr
		ClientType  string `json:"client_type,omitempty"`
		RedirectURI string `json:"redirect_uri,omitempty"`
	}{
		ofr:         ofr(r),
		ClientType:  r.clientType,
		RedirectURI: r.redirectURI,
	}
	data, err := json.Marshal(o)
	return data, errors.WithStack(err)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (r *OIDCFlowRequest) UnmarshalJSON(data []byte) error {
	type ofr OIDCFlowRequest
	o := struct {
		ofr
		RedirectType string `json:"client_type"`
		RedirectURI  string `json:"redirect_uri"`
	}{
		ofr: ofr(*NewOIDCFlowRequest()),
	}
	o.RedirectType = o.clientType
	o.RedirectURI = o.redirectURI
	if err := json.Unmarshal(data, &o); err != nil {
		return errors.WithStack(err)
	}
	o.clientType = o.RedirectType
	o.redirectURI = o.RedirectURI
	*r = OIDCFlowRequest(o.ofr)
	if r.SubtokenCapabilities != nil && !r.Capabilities.Has(api.CapabilityCreateMT) {
		r.SubtokenCapabilities = nil
	}
	return nil
}

// ToAuthCodeFlowRequest creates a AuthCodeFlowRequest from the OIDCFlowRequest
func (r OIDCFlowRequest) ToAuthCodeFlowRequest() AuthCodeFlowRequest {
	return AuthCodeFlowRequest{
		OIDCFlowRequest: r,
		ClientType:      r.clientType,
		RedirectURI:     r.redirectURI,
	}
}

// Scan implements the sql.Scanner interface
func (r *OIDCFlowRequest) Scan(src interface{}) error {
	v, ok := src.([]byte)
	if !ok {
		return errors.New("bad []byte type assertion")
	}
	return errors.WithStack(json.Unmarshal(v, r))
}

// Value implements the driver.Valuer interface
func (r OIDCFlowRequest) Value() (driver.Value, error) {
	v, err := json.Marshal(r)
	return v, errors.WithStack(err)
}
