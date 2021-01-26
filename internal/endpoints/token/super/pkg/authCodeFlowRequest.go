package pkg

import (
	"encoding/json"

	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/supertoken/capabilities"
	"github.com/oidc-mytoken/server/shared/supertoken/restrictions"
)

// Redirect types
const (
	redirectTypeWeb    = "web"
	redirectTypeNative = "native"
)

// AuthCodeFlowRequest holds a authorization code flow request
type AuthCodeFlowRequest struct {
	Issuer               string                    `json:"oidc_issuer"`
	GrantType            model.GrantType           `json:"grant_type"`
	OIDCFlow             model.OIDCFlow            `json:"oidc_flow"`
	Restrictions         restrictions.Restrictions `json:"restrictions"`
	Capabilities         capabilities.Capabilities `json:"capabilities"`
	SubtokenCapabilities capabilities.Capabilities `json:"subtoken_capabilities"`
	RedirectType         string                    `json:"redirect_type"`
	Name                 string                    `json:"name"`
	ResponseType         model.ResponseType        `json:"response_type"`
}

// NewAuthCodeFlowRequest creates a new AuthCodeFlowRequest with default values where they can be omitted
func NewAuthCodeFlowRequest() *AuthCodeFlowRequest {
	return &AuthCodeFlowRequest{
		RedirectType: redirectTypeWeb,
		Capabilities: capabilities.Capabilities{capabilities.CapabilityAT},
		ResponseType: model.ResponseTypeToken,
	}
}

// Native checks if the request is native
func (r *AuthCodeFlowRequest) Native() bool {
	if r.RedirectType == redirectTypeNative {
		return true
	}
	return false
}

// UnmarshalJSON implements the json unmarshaler interface
func (r *AuthCodeFlowRequest) UnmarshalJSON(data []byte) error {
	type authCodeFlowRequest2 AuthCodeFlowRequest
	rr := (*authCodeFlowRequest2)(NewAuthCodeFlowRequest())
	if err := json.Unmarshal(data, &rr); err != nil {
		return err
	}
	*r = AuthCodeFlowRequest(*rr)
	if r.SubtokenCapabilities != nil && !r.Capabilities.Has(capabilities.CapabilityCreateST) {
		r.SubtokenCapabilities = nil
	}
	return nil
}
