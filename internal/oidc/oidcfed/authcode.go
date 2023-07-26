package oidcfed

import (
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/oidc-mytoken/utils/utils"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/oidc/pkce"
	"github.com/oidc-mytoken/server/internal/server/routes"
)

// GetAuthorizationURL creates an authorization url using oidcfed automatic client registration
func (p OIDCFedProvider) GetAuthorizationURL(
	rlog log.Ext1FieldLogger, state, pkceChallenge string,
	scopeRestrictions, audRestrictions []string,
) (string, error) {
	rlog.Debug("Generating oidcfed authorization url")
	scopes := scopeRestrictions
	if len(scopes) <= 0 {
		scopes = p.Scopes()
	}
	if !utils.StringInSlice(oidc.ScopeOfflineAccess, scopes) {
		scopes = append(scopes, oidc.ScopeOfflineAccess)
	}
	// Even if user deselected openid scope in restriction, we still need it
	if !utils.StringInSlice(oidc.ScopeOpenID, scopes) {
		scopes = append(scopes, oidc.ScopeOpenID)
	}
	params := url.Values{}
	params.Set("prompt", "consent")
	params.Set("code_challenge", pkceChallenge)
	params.Set("code_challenge_method", pkce.TransformationS256.String())
	params.Set("nonce", utils.RandASCIIString(44)) // This is only here because some oidcfed implementations
	// require nonce, as we don't care about the id token we also don't check the nonce

	if len(audRestrictions) > 0 {
		params["resource"] = audRestrictions
	}

	return fedLeafEntity().GetAuthorizationURL(p.Issuer(), routes.RedirectURI, state, strings.Join(scopes, " "), params)
}
