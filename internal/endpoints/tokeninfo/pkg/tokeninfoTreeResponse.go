package pkg

import (
	"github.com/oidc-mytoken/server/internal/db/dbrepo/supertokenrepo/tree"
)

type TokeninfoTreeResponse struct {
	Tokens tree.SuperTokenEntryTree `json:"super_tokens"`
}

func NewTokeninfoTreeResponse(t tree.SuperTokenEntryTree) TokeninfoTreeResponse {
	return TokeninfoTreeResponse{Tokens: t}
}
