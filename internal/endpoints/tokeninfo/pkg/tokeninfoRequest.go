package pkg

import (
	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/mytoken/token"
)

type TokenInfoRequest struct {
	Action  model.TokeninfoAction `json:"action"`
	Mytoken token.Token           `json:"mytoken"`
}
