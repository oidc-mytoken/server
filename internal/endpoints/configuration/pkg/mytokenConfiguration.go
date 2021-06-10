package pkg

import (
	"github.com/oidc-mytoken/api/v0"
	model2 "github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/shared/model"
)

// MytokenConfiguration holds information about a mytoken instance
type MytokenConfiguration struct {
	api.MytokenConfiguration               `json:",inline"`
	TokenInfoEndpointActionsSupported      []model.TokeninfoAction `json:"tokeninfo_endpoint_actions_supported,omitempty"`
	AccessTokenEndpointGrantTypesSupported []model.GrantType       `json:"access_token_endpoint_grant_types_supported"`
	MytokenEndpointGrantTypesSupported     []model.GrantType       `json:"mytoken_endpoint_grant_types_supported"`
	MytokenEndpointOIDCFlowsSupported      []model.OIDCFlow        `json:"mytoken_endpoint_oidc_flows_supported"`
	ResponseTypesSupported                 []model.ResponseType    `json:"response_types_supported"`
	SupportedRestrictionKeys               model2.RestrictionKeys  `json:"supported_restriction_keys"`
	TokenEndpoint                          string                  `json:"token_endpoint"` // For compatibility with OIDC
}
