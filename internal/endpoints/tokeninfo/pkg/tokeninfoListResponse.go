package pkg

import (
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
)

// TokeninfoListResponse is type for responses to tokeninfo list requests
type TokeninfoListResponse struct {
	// un update check api.TokeninfoListResponse
	Tokens []tree.MytokenEntryTree `json:"mytokens"`
}

// NewTokeninfoListResponse creates a new TokeninfoListResponse
func NewTokeninfoListResponse(l []tree.MytokenEntryTree) TokeninfoListResponse {
	return TokeninfoListResponse{Tokens: l}
}
