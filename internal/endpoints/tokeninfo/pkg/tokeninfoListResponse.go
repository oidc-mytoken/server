package pkg

import (
	"github.com/oidc-mytoken/server/internal/db/dbrepo/supertokenrepo/tree"
)

type TokeninfoListResponse struct {
	Tokens []tree.SuperTokenEntryTree `json:"super_tokens"`
}

func NewTokeninfoListResponse(l []tree.SuperTokenEntryTree) TokeninfoListResponse {
	return TokeninfoListResponse{Tokens: l}
}
