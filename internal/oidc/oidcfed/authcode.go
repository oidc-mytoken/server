package oidcfed

import (
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	oidcfed "github.com/zachmann/go-oidcfed/pkg"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/jws"
	"github.com/oidc-mytoken/server/internal/oidc/pkce"
	"github.com/oidc-mytoken/server/internal/server/routes"
)

var requestObjectProducer *oidcfed.RequestObjectProducer

func InitOIDCFedAuthCode() {
	jws.LoadOIDCSigningKey()
	requestObjectProducer = oidcfed.NewRequestObjectProducer(
		config.Get().IssuerURL,
		jws.GetSigningKey(jws.KeyUsageOIDCSigning), config.Get().Signing.OIDC.Alg, 60,
	)
}

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
	requestParams := map[string]any{
		"aud":                   p.Issuer(),
		"redirect_uri":          routes.RedirectURI,
		"prompt":                "consent",
		"code_challenge":        pkceChallenge,
		"code_challenge_method": pkce.TransformationS256.String(),
		"state":                 state,
		"response_type":         "code",
		"scope":                 scopes,
		"nonce":                 utils.RandASCIIString(44), // This is only here because some oidcfed implementations
		// require nonce, as we don't care about the id token we also don't check the nonce
	}
	auds := audRestrictions
	if len(auds) > 0 {
		requestParams["resource"] = auds
	}

	requestObject, err := requestObjectProducer.RequestObject(requestParams)
	if err != nil {
		return "", errors.Wrap(err, "could not create request object")
	}
	u, err := url.Parse(p.Endpoints().Authorization)
	if err != nil {
		return "", errors.WithStack(err)
	}
	q := url.Values{}
	q.Set("request", string(requestObject))
	q.Set("client_id", p.ClientID())
	q.Set("response_type", "code")
	q.Set("redirect_uri", routes.RedirectURI)
	q.Set("scope", strings.Join(scopes, " "))
	u.RawQuery = q.Encode()
	return u.String(), nil
}
