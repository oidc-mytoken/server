package pkg

import (
	"github.com/oidc-mytoken/server/internal/db/dbrepo/eventrepo"
	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
)

// TokeninfoHistoryResponse is type for responses to tokeninfo history requests
type TokeninfoHistoryResponse struct {
	// un update check api.TokeninfoHistoryResponse
	EventHistory eventrepo.EventHistory `json:"events"`
	TokenUpdate  *my.MytokenResponse    `json:"token_update,omitempty"`
}

// NewTokeninfoHistoryResponse creates a new TokeninfoHistoryResponse
func NewTokeninfoHistoryResponse(h eventrepo.EventHistory) TokeninfoHistoryResponse {
	return TokeninfoHistoryResponse{EventHistory: h}
}
