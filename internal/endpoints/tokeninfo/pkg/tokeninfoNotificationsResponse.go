package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
)

// TokeninfoNotificationsResponse is a type for responses to tokeninfo notification requests
type TokeninfoNotificationsResponse struct {
	// on update check api.TokeninfoNotificationsResponse
	api.TokeninfoNotificationsResponse
	TokenUpdate *my.MytokenResponse `json:"token_update,omitempty"`
}
