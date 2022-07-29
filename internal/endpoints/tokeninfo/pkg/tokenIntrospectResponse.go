package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/shared/model"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
)

// TokeninfoIntrospectResponse is type for responses to tokeninfo introspect requests
type TokeninfoIntrospectResponse struct {
	api.TokeninfoIntrospectResponse `json:",inline"`
	TokenType                       model.ResponseType  `json:"token_type"`
	Token                           mytoken.UsedMytoken `json:"token"`
}
