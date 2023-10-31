package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
)

// MytokenResponse is a response to a mytoken request
type MytokenResponse struct {
	api.MytokenResponse `json:",inline"`
	MytokenType         model.ResponseType        `json:"mytoken_type"`
	Restrictions        restrictions.Restrictions `json:"restrictions,omitempty"`
	TokenUpdate         *MytokenResponse          `json:"token_update,omitempty"`
}

// OnlyTokenUpdateRes is a response that contains only a TokenUpdate and is used when a rotating mytoken was used but
// no response is returned otherwise
type OnlyTokenUpdateRes struct {
	api.OnlyTokenUpdateResponse
	TokenUpdate *MytokenResponse `json:"token_update,omitempty"`
}

// SetTokenUpdate implements the pkg.TokenUpdatableResponse interface
func (res *OnlyTokenUpdateRes) SetTokenUpdate(tokenUpdate *MytokenResponse) {
	res.TokenUpdate = tokenUpdate
}
