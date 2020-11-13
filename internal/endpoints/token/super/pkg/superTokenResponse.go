package pkg

import (
	"github.com/zachmann/mytoken/internal/supertoken/capabilities"
	"github.com/zachmann/mytoken/internal/supertoken/restrictions"
)

type SuperTokenResponse struct {
	SuperToken   string                    `json:"super_token"`
	ExpiresIn    uint64                    `json:"expires_in,omitempty"`
	Restrictions restrictions.Restrictions `json:"restrictions,omitempty"`
	Capabilities capabilities.Capabilities `json:"capabilities,omitempty"`
}
