package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
)

// TokenInfoRequest is a type for holding a request to the tokeninfo endpoint
type TokenInfoRequest struct {
	api.TokenInfoRequest
	Action  model.TokeninfoAction             `json:"action"`
	Mytoken universalmytoken.UniversalMytoken `json:"mytoken"`
}
