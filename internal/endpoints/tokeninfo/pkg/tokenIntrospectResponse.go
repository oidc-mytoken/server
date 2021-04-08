package pkg

import (
	"github.com/oidc-mytoken/server/pkg/api/v0"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
)

type TokeninfoIntrospectResponse struct {
	api.TokeninfoIntrospectResponse `json:",inline"`
	Token                           mytoken.UsedMytoken `json:"token"`
}
