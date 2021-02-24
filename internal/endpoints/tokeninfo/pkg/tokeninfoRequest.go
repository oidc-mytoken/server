package pkg

import (
	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/supertoken/token"
)

type TokenInfoRequest struct {
	Action     model.TokeninfoAction `json:"action"`
	SuperToken token.Token           `json:"super_token"`
}
