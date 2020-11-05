package pkg

import (
	"github.com/zachmann/mytoken/internal/supertoken/capabilities"
	"github.com/zachmann/mytoken/internal/supertoken/restrictions"
)

type SuperTokenResponse struct {
	SuperToken   string                    `json:"super_token"`
	ExpiresIn    uint64                    `json:"expires_in"`
	Restrictions restrictions.Restrictions `json:"restrictions"`
	Capabilities capabilities.Capabilities `json:"capabilities"`
}
