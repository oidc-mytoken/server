package pkg

import (
	"github.com/oidc-mytoken/server/internal/model/profiled"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
)

// MytokenFromMytokenRequest is a request to create a new Mytoken from an existing Mytoken
// It is analog to api.MytokenFromMytokenRequest
type MytokenFromMytokenRequest struct {
	profiled.GeneralMytokenRequest `json:",inline"`
	Mytoken                        universalmytoken.UniversalMytoken `json:"mytoken"`
	FailOnRestrictionsNotTighter   bool                              `json:"error_on_restrictions,omitempty"`
}

// NewMytokenRequest creates a MytokenFromMytokenRequest with the default values where they can be omitted
func NewMytokenRequest() *MytokenFromMytokenRequest {
	return &MytokenFromMytokenRequest{
		GeneralMytokenRequest: *profiled.NewGeneralMytokenRequest(),
	}
}

/*
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
*/
