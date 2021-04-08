package pkg

import (
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
)

type TokeninfoTreeResponse struct {
	// un update check api.TokeninforTeeResponse
	Tokens tree.MytokenEntryTree `json:"mytokens"`
}

func NewTokeninfoTreeResponse(t tree.MytokenEntryTree) TokeninfoTreeResponse {
	return TokeninfoTreeResponse{Tokens: t}
}
