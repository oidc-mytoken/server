package pkg

import (
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
)

type TokeninfoListResponse struct {
	Tokens []tree.MytokenEntryTree `json:"mytokens"`
}

func NewTokeninfoListResponse(l []tree.MytokenEntryTree) TokeninfoListResponse {
	return TokeninfoListResponse{Tokens: l}
}
