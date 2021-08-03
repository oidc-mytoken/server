package pkg

import (
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
)

// ConsentPostRequest holds the post request for confirming consent
type ConsentPostRequest struct {
	Issuer               string                    `json:"oidc_iss"`
	Restrictions         restrictions.Restrictions `json:"restrictions"`
	Capabilities         api.Capabilities          `json:"capabilities"`
	SubtokenCapabilities api.Capabilities          `json:"subtoken_capabilities"`
	TokenName            string                    `json:"name"`
	Rotation             *api.Rotation             `json:"rotation,omitempty"`
}
