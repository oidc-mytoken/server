package pkg

import (
	"encoding/json"

	"github.com/zachmann/mytoken/internal/model"

	"github.com/zachmann/mytoken/internal/server/supertoken/capabilities"
	"github.com/zachmann/mytoken/internal/server/supertoken/restrictions"
	"github.com/zachmann/mytoken/internal/server/supertoken/token"
)

// SuperTokenFromSuperTokenRequest is a request to create a new supertoken from an existing supertoken
type SuperTokenFromSuperTokenRequest struct {
	Issuer               string                    `json:"oidc_issuer"`
	GrantType            model.GrantType           `json:"grant_type"`
	SuperToken           token.Token               `json:"super_token"`
	Restrictions         restrictions.Restrictions `json:"restrictions"`
	Capabilities         capabilities.Capabilities `json:"capabilities"`
	SubtokenCapabilities capabilities.Capabilities `json:"subtoken_capabilities"`
	Name                 string                    `json:"name"`
	ResponseType         model.ResponseType        `json:"response_type"`
}

// NewSuperTokenRequest creates a SuperTokenFromSuperTokenRequest with the default values where they can be omitted
func NewSuperTokenRequest() *SuperTokenFromSuperTokenRequest {
	return &SuperTokenFromSuperTokenRequest{
		ResponseType: model.ResponseTypeToken,
	}
}

// UnmarshalJSON implements the json unmarshaler interface
func (r *SuperTokenFromSuperTokenRequest) UnmarshalJSON(data []byte) error {
	type superTokenFromSuperTokenRequest2 SuperTokenFromSuperTokenRequest
	rr := (*superTokenFromSuperTokenRequest2)(NewSuperTokenRequest())
	if err := json.Unmarshal(data, &rr); err != nil {
		return err
	}
	*r = SuperTokenFromSuperTokenRequest(*rr)
	if r.SubtokenCapabilities != nil && !r.Capabilities.Has(capabilities.CapabilityCreateST) {
		r.SubtokenCapabilities = nil
	}
	return nil
}
