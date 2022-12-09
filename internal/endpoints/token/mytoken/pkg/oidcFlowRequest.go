package pkg

import (
	"database/sql/driver"
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/model/profiled"
)

// OIDCFlowRequest holds the request for an OIDC Flow request
type OIDCFlowRequest struct {
	profiled.GeneralMytokenRequest
	OIDCFlowAttrs
}

// OIDCFlowAttrs holds the additional attributes for OIDCFlowRequests
type OIDCFlowAttrs struct {
	OIDCFlow model.OIDCFlow `json:"oidc_flow"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (r *OIDCFlowRequest) UnmarshalJSON(data []byte) error {
	if err := errors.WithStack(json.Unmarshal(data, &r.GeneralMytokenRequest)); err != nil {
		return err
	}
	return errors.WithStack(json.Unmarshal(data, &r.OIDCFlowAttrs))
}

// MarshalJSON implements the json.Marshaler interface
func (r OIDCFlowRequest) MarshalJSON() ([]byte, error) {
	base, err := json.Marshal(r.GeneralMytokenRequest)
	if err != nil {
		return nil, err
	}
	attrs, err := json.Marshal(r.OIDCFlowAttrs)
	if err != nil {
		return nil, err
	}
	return jsonpatch.MergePatch(base, attrs)
}

// NewOIDCFlowRequest creates a new OIDCFlowRequest with default values where they can be omitted
func NewOIDCFlowRequest() *OIDCFlowRequest {
	return &OIDCFlowRequest{
		GeneralMytokenRequest: *profiled.NewGeneralMytokenRequest(),
		// clientType:            api.ClientTypeNative,
	}
}

/*
// SetRedirectType sets the (hidden) redirect type
func (r *OIDCFlowRequest) SetRedirectType(redirect string) {
	r.clientType = redirect
}

// SetRedirectURI sets the (hidden) redirect uri
func (r *OIDCFlowRequest) SetRedirectURI(uri string) {
	r.redirectURI = uri
}

// MarshalJSON implements the json.Marshaler interface
func (r OIDCFlowRequest) MarshalJSON() ([]byte, error) { // skipcq: CRT-P0003
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
	return nil
}

// ToAuthCodeFlowRequest creates a AuthCodeFlowRequest from the OIDCFlowRequest
func (r OIDCFlowRequest) ToAuthCodeFlowRequest() AuthCodeFlowRequest { // skipcq: CRT-P0003
	return AuthCodeFlowRequest{
		OIDCFlowRequest: r,
		ClientType:      r.clientType,
		RedirectURI:     r.redirectURI,
	}
}
*/
// Scan implements the sql.Scanner interface
func (r *OIDCFlowRequest) Scan(src interface{}) error {
	v, ok := src.([]byte)
	if !ok {
		return errors.New("bad []byte type assertion")
	}
	return errors.WithStack(json.Unmarshal(v, r))
}

// Value implements the driver.Valuer interface
func (r OIDCFlowRequest) Value() (driver.Value, error) { // skipcq: CRT-P0003
	v, err := json.Marshal(r)
	return v, errors.WithStack(err)
}
