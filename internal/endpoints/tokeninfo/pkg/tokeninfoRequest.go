package pkg

import (
	"github.com/oidc-mytoken/server/pkg/api/v0"
	"github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/mytoken/token"
)

type TokenInfoRequest struct {
	api.TokenInfoRequest `json:",inline"`
	Action               model.TokeninfoAction `json:"action"`
	Mytoken              token.Token           `json:"mytoken"`
}
