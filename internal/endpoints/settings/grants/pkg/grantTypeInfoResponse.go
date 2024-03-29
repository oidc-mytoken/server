package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
)

// GrantTypeInfoResponse is a type for the response for listing grant types
type GrantTypeInfoResponse struct {
	api.GrantTypeInfoResponse
	TokenUpdate *my.MytokenResponse `json:"token_update,omitempty"`
}

// SetTokenUpdate implements the pkg.TokenUpdatableResponse interface
func (res *GrantTypeInfoResponse) SetTokenUpdate(tokenUpdate *my.MytokenResponse) {
	res.TokenUpdate = tokenUpdate
}
