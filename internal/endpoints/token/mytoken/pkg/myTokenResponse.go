package pkg

import (
	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/mytoken/capabilities"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
)

// MytokenResponse is a response to a mytoken request
type MytokenResponse struct {
	Mytoken              string                    `json:"mytoken,omitempty"`
	MytokenType          model.ResponseType        `json:"mytoken_type"`
	TransferCode         string                    `json:"transfer_code,omitempty"`
	ExpiresIn            uint64                    `json:"expires_in,omitempty"`
	Restrictions         restrictions.Restrictions `json:"restrictions,omitempty"`
	Capabilities         capabilities.Capabilities `json:"capabilities,omitempty"`
	SubtokenCapabilities capabilities.Capabilities `json:"subtoken_capabilities,omitempty"`
}
