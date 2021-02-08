package pkg

import (
	"github.com/oidc-mytoken/server/shared/supertoken/capabilities"
	"github.com/oidc-mytoken/server/shared/supertoken/restrictions"
)

type ConsentPostRequest struct {
	Restrictions         restrictions.Restrictions `json:"restrictions"`
	Capabilities         capabilities.Capabilities `json:"capabilities"`
	SubtokenCapabilities capabilities.Capabilities `json:"subtoken_capabilities"`
}
