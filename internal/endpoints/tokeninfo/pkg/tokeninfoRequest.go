package pkg

import (
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/mytoken/token"
)

// TokenInfoRequest is a type for holding a request to the tokeninfo endpoint
type TokenInfoRequest struct {
	api.TokenInfoRequest `json:",inline"`
	Action               model.TokeninfoAction `json:"action"`
	Mytoken              token.Token           `json:"mytoken"`
}
