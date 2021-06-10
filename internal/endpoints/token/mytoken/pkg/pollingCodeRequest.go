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
