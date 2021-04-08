package mytokenlib

import (
	"strings"

	api "github.com/oidc-mytoken/server/pkg/api/v0"
)

func NewAccessTokenRequest(issuer, mytoken string, scopes, audiences []string, comment string) api.AccessTokenRequest {
	return api.AccessTokenRequest{
		Issuer:    issuer,
		GrantType: api.GrantTypeMytoken,
		Scope:     strings.Join(scopes, " "),
		Audience:  strings.Join(audiences, " "),
		Comment:   comment,
		Mytoken:   mytoken,
	}
}
