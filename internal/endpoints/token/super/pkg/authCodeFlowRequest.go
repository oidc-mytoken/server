package pkg

import (
	"encoding/json"

	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/supertoken/capabilities"
	"github.com/zachmann/mytoken/internal/supertoken/restrictions"
)

// Redirect types
const (
	RedirectTypeWeb    = "web"
	RedirectTypeNative = "native"
)

type AuthCodeFlowRequest struct {
	Issuer               string                    `json:"oidc_issuer"`
	GrantType            model.GrantType           `json:"grant_type"`
	OIDCFlow             model.OIDCFlow            `json:"oidc_flow"`
	Restrictions         restrictions.Restrictions `json:"restrictions"`
	Capabilities         capabilities.Capabilities `json:"capabilities"`
	SubtokenCapabilities capabilities.Capabilities `json:"subtoken_capabilities"`
	RedirectType         string                    `json:"redirect_type"`
	Name                 string                    `json:"name"`
}

func NewAuthCodeFlowRequest() *AuthCodeFlowRequest {
	return &AuthCodeFlowRequest{
		RedirectType: RedirectTypeWeb,
		Capabilities: capabilities.Capabilities{capabilities.CapabilityAT},
	}
}

func (r *AuthCodeFlowRequest) Native() bool {
	if r.RedirectType == RedirectTypeNative {
		return true
	}
	return false
}

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
