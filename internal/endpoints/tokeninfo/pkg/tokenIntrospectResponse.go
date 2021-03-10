package pkg

import (
	supertoken "github.com/oidc-mytoken/server/shared/supertoken/pkg"
)

type TokeninfoIntrospectResponse struct {
	Valid bool                      `json:"valid"`
	Token supertoken.UsedSuperToken `json:"token"`
}
