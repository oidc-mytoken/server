package pkg

import (
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
)

type TokeninfoListResponse struct {
	// un update check api.TokeninfoListResponse
	Tokens []tree.MytokenEntryTree `json:"mytokens"`
}

func NewTokeninfoListResponse(l []tree.MytokenEntryTree) TokeninfoListResponse {
	return TokeninfoListResponse{Tokens: l}
}
