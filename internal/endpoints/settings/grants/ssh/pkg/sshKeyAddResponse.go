package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
)

// SSHKeyAddResponse is a type for the (first) response to an SSHKeyAddRequest
type SSHKeyAddResponse struct {
	api.SSHKeyAddResponse
	TokenUpdate *my.MytokenResponse `json:"token_update,omitempty"`
}

// SetTokenUpdate implements the pkg.TokenUpdatableResponse interface
func (info *SSHKeyAddResponse) SetTokenUpdate(tokenUpdate *my.MytokenResponse) {
	info.TokenUpdate = tokenUpdate
}
