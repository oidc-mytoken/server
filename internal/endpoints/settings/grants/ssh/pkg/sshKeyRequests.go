package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
	"github.com/oidc-mytoken/server/shared/mytoken/universalmytoken"
)

type SSHKeyDeleteRequest struct {
	api.SSHKeyDeleteRequest
	Mytoken universalmytoken.UniversalMytoken `json:"mytoken" xml:"mytoken" form:"mytoken"`
}

type SSHKeyAddRequest struct {
	api.SSHKeyAddRequest
	Mytoken      universalmytoken.UniversalMytoken `json:"mytoken" xml:"mytoken" form:"mytoken"`
	Restrictions restrictions.Restrictions         `json:"restrictions" form:"restrictions" xml:"restrictions"`
}
