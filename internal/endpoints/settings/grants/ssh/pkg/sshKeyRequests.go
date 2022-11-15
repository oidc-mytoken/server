package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
)

// SSHKeyDeleteRequest is a type for a request to delete an ssh key
type SSHKeyDeleteRequest struct {
	api.SSHKeyDeleteRequest
	Mytoken universalmytoken.UniversalMytoken `json:"mytoken" xml:"mytoken" form:"mytoken"`
}

// SSHKeyAddRequest is a type for a request to add an ssh key
type SSHKeyAddRequest struct {
	api.SSHKeyAddRequest
	Mytoken      universalmytoken.UniversalMytoken `json:"mytoken" xml:"mytoken" form:"mytoken"`
	Restrictions restrictions.Restrictions         `json:"restrictions" form:"restrictions" xml:"restrictions"`
}
