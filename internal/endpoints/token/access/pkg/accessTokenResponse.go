package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
)

// AccessTokenResponse is type for responses for access token requests
type AccessTokenResponse struct {
	// un update check api.AccessTokenResponse
	api.AccessTokenResponse
	TokenUpdate *my.MytokenResponse `json:"token_update,omitempty"`
}
