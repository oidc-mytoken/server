package pkg

import (
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
)

// TokeninfoTreeResponse is type for responses to tokeninfo tree requests
type TokeninfoTreeResponse struct {
	// on update check api.TokeninfoTreeResponse
	Tokens      tree.MytokenEntryTree `json:"mytokens"`
	TokenUpdate *my.MytokenResponse   `json:"token_update,omitempty"`
}

// NewTokeninfoTreeResponse creates a new TokeninfoTreeResponse
func NewTokeninfoTreeResponse(t tree.MytokenEntryTree, update *my.MytokenResponse) TokeninfoTreeResponse {
	return TokeninfoTreeResponse{Tokens: t, TokenUpdate: update}
}
