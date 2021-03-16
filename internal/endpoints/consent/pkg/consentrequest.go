package pkg

import (
	"github.com/oidc-mytoken/server/shared/mytoken/capabilities"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
)

// ConsentPostRequest holds the post request for confirming consent
type ConsentPostRequest struct {
	Issuer               string                    `json:"oidc_iss"`
	Restrictions         restrictions.Restrictions `json:"restrictions"`
	Capabilities         capabilities.Capabilities `json:"capabilities"`
	SubtokenCapabilities capabilities.Capabilities `json:"subtoken_capabilities"`
}
