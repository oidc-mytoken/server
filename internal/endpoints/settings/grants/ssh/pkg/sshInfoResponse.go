package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
)

type SSHInfoResponse struct {
	api.SSHInfoResponse
	TokenUpdate *my.MytokenResponse `json:"token_update,omitempty"`
}

// SetTokenUpdate implements the pkg.TokenUpdatableResponse interface
func (info *SSHInfoResponse) SetTokenUpdate(tokenUpdate *my.MytokenResponse) {
	info.TokenUpdate = tokenUpdate
}
