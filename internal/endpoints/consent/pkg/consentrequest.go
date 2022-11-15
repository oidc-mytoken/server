package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
)

// ConsentApprovalRequest holds the post request for confirming consent
type ConsentApprovalRequest struct {
	Issuer       string                    `json:"oidc_issuer"`
	Restrictions restrictions.Restrictions `json:"restrictions"`
	Capabilities api.Capabilities          `json:"capabilities"`
	TokenName    string                    `json:"name"`
	Rotation     *api.Rotation             `json:"rotation,omitempty"`
}

// ConsentRequest holds the post request for creating a consent screen
type ConsentRequest struct {
	ConsentApprovalRequest
	ApplicationName string                            `json:"application_name"`
	Mytoken         universalmytoken.UniversalMytoken `json:"mytoken"`
}
