package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
)

// GrantTypeRequest is a type for a enable/disable grant type request
type GrantTypeRequest struct {
	api.GrantTypeRequest
	GrantType model.GrantType                   `json:"grant_type" xml:"grant_type" form:"grant_type"`
	Mytoken   universalmytoken.UniversalMytoken `json:"mytoken" xml:"mytoken" form:"mytoken"`
}
