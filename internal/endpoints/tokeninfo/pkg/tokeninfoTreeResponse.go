package pkg

import (
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
)

// TokeninfoSubtokensResponse is type for responses to tokeninfo tree requests
type TokeninfoSubtokensResponse struct {
	// on update check api.TokeninfoTreeResponse
	Tokens      tree.MytokenEntryTree `json:"mytokens"`
	TokenUpdate *my.MytokenResponse   `json:"token_update,omitempty"`
}

// NewTokeninfoSubtokensResponse creates a new TokeninfoSubtokensResponse
func NewTokeninfoSubtokensResponse(t tree.MytokenEntryTree, update *my.MytokenResponse) TokeninfoSubtokensResponse {
	return TokeninfoSubtokensResponse{
		Tokens:      t,
		TokenUpdate: update,
	}
}
