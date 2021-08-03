package pkg

import (
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
)

// TokeninfoListResponse is type for responses to tokeninfo list requests
type TokeninfoListResponse struct {
	// un update check api.TokeninfoListResponse
	Tokens      []tree.MytokenEntryTree `json:"mytokens"`
	TokenUpdate *my.MytokenResponse     `json:"token_update,omitempty"`
}

// NewTokeninfoListResponse creates a new TokeninfoListResponse
func NewTokeninfoListResponse(l []tree.MytokenEntryTree, update *my.MytokenResponse) TokeninfoListResponse {
	return TokeninfoListResponse{Tokens: l, TokenUpdate: update}
}
