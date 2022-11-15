package pkg

import (
	"encoding/json"

	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
)

// MytokenFromMytokenRequest is a request to create a new Mytoken from an existing Mytoken
type MytokenFromMytokenRequest struct {
	api.MytokenFromMytokenRequest `json:",inline"`
	GrantType                     model.GrantType                   `json:"grant_type"`
	Mytoken                       universalmytoken.UniversalMytoken `json:"mytoken"`
	Restrictions                  restrictions.Restrictions         `json:"restrictions"`
	ResponseType                  model.ResponseType                `json:"response_type"`
}

// NewMytokenRequest creates a MytokenFromMytokenRequest with the default values where they can be omitted
func NewMytokenRequest() *MytokenFromMytokenRequest {
	return &MytokenFromMytokenRequest{
		ResponseType: model.ResponseTypeToken,
		GrantType:    -1,
	}
}

// UnmarshalJSON implements the json unmarshaler interface
func (r *MytokenFromMytokenRequest) UnmarshalJSON(data []byte) error {
	type mytokenFromMytokenRequest2 MytokenFromMytokenRequest
	rr := (*mytokenFromMytokenRequest2)(NewMytokenRequest())
	if err := json.Unmarshal(data, &rr); err != nil {
		return errors.WithStack(err)
	}
	*r = MytokenFromMytokenRequest(*rr)
	return nil
}
