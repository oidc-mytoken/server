package pkg

import (
	"github.com/oidc-mytoken/server/internal/db/dbrepo/eventrepo"
)

type TokeninfoHistoryResponse struct {
	EventHistory eventrepo.EventHistory `json:"events"`
}

func NewTokeninfoHistoryResponse(h eventrepo.EventHistory) TokeninfoHistoryResponse {
	return TokeninfoHistoryResponse{EventHistory: h}
}
