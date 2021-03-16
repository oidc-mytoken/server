package pkg

import (
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
)

type TokeninfoIntrospectResponse struct {
	Valid bool                `json:"valid"`
	Token mytoken.UsedMytoken `json:"token"`
}
