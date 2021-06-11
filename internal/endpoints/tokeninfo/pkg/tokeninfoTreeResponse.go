package pkg

import (
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
)

// TokeninfoTreeResponse is type for responses to tokeninfo tree requests
type TokeninfoTreeResponse struct {
	// un update check api.TokeninforTeeResponse
	Tokens tree.MytokenEntryTree `json:"mytokens"`
}

// NewTokeninfoTreeResponse creates a new TokeninfoTreeResponse
func NewTokeninfoTreeResponse(t tree.MytokenEntryTree) TokeninfoTreeResponse {
	return TokeninfoTreeResponse{Tokens: t}
}
