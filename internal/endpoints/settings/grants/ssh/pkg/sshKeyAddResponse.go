package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
)

// SSHKeyAddResponse is a type for the (first) response to an SSHKeyAddRequest
type SSHKeyAddResponse struct {
	api.AuthCodeFlowResponse
	TokenUpdate *my.MytokenResponse `json:"token_update,omitempty"`
}

// SetTokenUpdate implements the pkg.TokenUpdatableResponse interface
func (info *SSHKeyAddResponse) SetTokenUpdate(tokenUpdate *my.MytokenResponse) {
	info.TokenUpdate = tokenUpdate
}

// SSHKeyAddFinalResponse is a type for the final response for an SSHKeyAddRequest after the polling was successful
type SSHKeyAddFinalResponse struct {
	SSHUser       string `json:"ssh_user"`
	SSHHostConfig string `json:"ssh_host_config,omitempty"`
}
