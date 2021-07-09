package pkg

import (
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/server/shared/model"
)

// PollingCodeRequest is a polling code request
type PollingCodeRequest struct {
	api.PollingCodeRequest `json:",inline"`
	GrantType              model.GrantType `json:"grant_type"`
}

// NewPollingCodeRequest returns a new PollingCodeRequest
func NewPollingCodeRequest() PollingCodeRequest {
	return PollingCodeRequest{
		GrantType: -1, // This value will remain if grant_type is not contained in the request. We have to set it to -1, because the default of 0 would be a valid GrantType
	}
}
