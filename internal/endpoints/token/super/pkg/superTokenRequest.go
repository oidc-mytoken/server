package pkg

import (
	"encoding/json"

	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/supertoken/capabilities"
	"github.com/zachmann/mytoken/internal/supertoken/restrictions"
)

type SuperTokenFromSuperTokenRequest struct {
	Issuer               string                    `json:"oidc_issuer"`
	GrantType            model.GrantType           `json:"grant_type"`
	SuperToken           string                    `json:"super_token"`
	Restrictions         restrictions.Restrictions `json:"restrictions"`
	Capabilities         capabilities.Capabilities `json:"capabilities"`
	SubtokenCapabilities capabilities.Capabilities `json:"subtoken_capabilities"`
	Name                 string                    `json:"name"`
	ResponseType         model.ResponseType        `json:"response_type"`
}

func NewSuperTokenRequest() *SuperTokenFromSuperTokenRequest {
	return &SuperTokenFromSuperTokenRequest{
		ResponseType: model.ResponseTypeToken,
	}
}

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
