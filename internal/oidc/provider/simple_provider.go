package provider

import (
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-resty/resty/v2"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/oidc-mytoken/utils/utils/issuerutils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/issuer"
	"github.com/oidc-mytoken/server/internal/oidc/pkce"
	"github.com/oidc-mytoken/server/internal/server/routes"
	"github.com/oidc-mytoken/server/pkg/oauth2x"
)

type SimpleProvider struct {
	config.ProviderConf
}

func (p SimpleProvider) Name() string {
	return p.ProviderConf.Name
}

func (p SimpleProvider) Issuer() string {
	return p.ProviderConf.Issuer
}

func (p SimpleProvider) ClientID() string {
	return p.ProviderConf.ClientID
}

func (p SimpleProvider) Scopes() []string {
	return p.ProviderConf.Scopes
}

func (p SimpleProvider) Endpoints() *oauth2x.Endpoints {
	return p.ProviderConf.Endpoints
}

func (p SimpleProvider) Audience() *model.AudienceConf {
	return p.ProviderConf.Audience
}

func (p SimpleProvider) MaxMytokenLifetime() int64 {
	return p.MytokensMaxLifetime
}

func (p SimpleProvider) AddClientAuthentication(r *resty.Request, _ string) *resty.Request {
	return r.SetBasicAuth(p.ClientID(), p.ClientSecret)
}

// GetAuthorizationURL creates an authorization url
func (p SimpleProvider) GetAuthorizationURL(
	rlog log.Ext1FieldLogger, state, pkceChallenge string,
	scopeRestrictions, audRestrictions []string,
) (string, error) {
	rlog.Debug("Generating authorization url")
	scopes := scopeRestrictions
	if len(scopes) <= 0 {
		scopes = p.Scopes()
	}
	oauth2Config := oauth2.Config{
		ClientID:     p.ClientID(),
		ClientSecret: p.ClientSecret,
		Endpoint:     p.Endpoints().OAuth2(),
		RedirectURL:  routes.RedirectURI,
		Scopes:       scopes,
	}
	additionalParams := []oauth2.AuthCodeOption{
		oauth2.ApprovalForce,
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", pkce.TransformationS256.String()),
	}
	if issuerutils.CompareIssuerURLs(p.Issuer(), issuer.GOOGLE) {
		additionalParams = append(additionalParams, oauth2.AccessTypeOffline)
	} else if !utils.StringInSlice(oidc.ScopeOfflineAccess, oauth2Config.Scopes) {
		oauth2Config.Scopes = append(oauth2Config.Scopes, oidc.ScopeOfflineAccess)
	}
	// Even if user deselected openid scope in restriction, we still need it
	if !utils.StringInSlice(oidc.ScopeOpenID, oauth2Config.Scopes) {
		oauth2Config.Scopes = append(oauth2Config.Scopes, oidc.ScopeOpenID)
	}
	auds := audRestrictions
	if len(auds) > 0 {
		if p.Audience().SpaceSeparateAuds {
			additionalParams = append(
				additionalParams,
				oauth2.SetAuthURLParam(p.Audience().RequestParameter, strings.Join(auds, " ")),
			)
		} else {
			for _, a := range auds {
				additionalParams = append(
					additionalParams, oauth2.SetAuthURLParam(p.Audience().RequestParameter, a),
				)
			}
		}
	}

	return oauth2Config.AuthCodeURL(state, additionalParams...), nil
}
