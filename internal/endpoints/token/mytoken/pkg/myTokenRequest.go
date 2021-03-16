package pkg

import (
	"encoding/json"

	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/mytoken/capabilities"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
	"github.com/oidc-mytoken/server/shared/mytoken/token"
)

// MytokenFromMytokenRequest is a request to create a new Mytoken from an existing Mytoken
type MytokenFromMytokenRequest struct {
	Issuer               string                    `json:"oidc_issuer"`
	GrantType            model.GrantType           `json:"grant_type"`
	Mytoken              token.Token               `json:"mytoken"`
	Restrictions         restrictions.Restrictions `json:"restrictions"`
	Capabilities         capabilities.Capabilities `json:"capabilities"`
	SubtokenCapabilities capabilities.Capabilities `json:"subtoken_capabilities"`
	Name                 string                    `json:"name"`
	ResponseType         model.ResponseType        `json:"response_type"`
}

// NewMytokenRequest creates a MytokenFromMytokenRequest with the default values where they can be omitted
func NewMytokenRequest() *MytokenFromMytokenRequest {
	return &MytokenFromMytokenRequest{
		ResponseType: model.ResponseTypeToken,
	}
}

// UnmarshalJSON implements the json unmarshaler interface
func (r *MytokenFromMytokenRequest) UnmarshalJSON(data []byte) error {
	type mytokenFromMytokenRequest2 MytokenFromMytokenRequest
	rr := (*mytokenFromMytokenRequest2)(NewMytokenRequest())
	if err := json.Unmarshal(data, &rr); err != nil {
		return err
	}
	*r = MytokenFromMytokenRequest(*rr)
	if r.SubtokenCapabilities != nil && !r.Capabilities.Has(capabilities.CapabilityCreateMT) {
		r.SubtokenCapabilities = nil
	}
	return nil
}
