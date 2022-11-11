package routes

import (
	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/server/apipath"
	"github.com/oidc-mytoken/server/shared/utils"
)

var routes *paths

// WellknownMytokenConfiguration is the mytoken configuration path suffix
const WellknownMytokenConfiguration = "/.well-known/mytoken-configuration"

// WellknownOpenIDConfiguration is the openid configuration path suffix
const WellknownOpenIDConfiguration = "/.well-known/openid-configuration"

// init initializes the server route paths
func init() {
	routes = &paths{
		api: map[int]APIPaths{
			0: {
				MytokenEndpoint:       utils.CombineURLPath(apipath.V0, "/token/my"),
				AccessTokenEndpoint:   utils.CombineURLPath(apipath.V0, "/token/access"),
				TokenInfoEndpoint:     utils.CombineURLPath(apipath.V0, "/tokeninfo"),
				RevocationEndpoint:    utils.CombineURLPath(apipath.V0, "/token/revoke"),
				TokenTransferEndpoint: utils.CombineURLPath(apipath.V0, "/token/transfer"),
				UserSettingEndpoint:   utils.CombineURLPath(apipath.V0, "/settings"),
			},
		},
		other: GeneralPaths{
			ConfigurationEndpoint: WellknownMytokenConfiguration,
			OIDCRedirectEndpoint:  "/redirect",
			JWKSEndpoint:          "/jwks",
			ConsentEndpoint:       "/c",
			Privacy:               "/privacy",
		},
	}
}

type paths struct {
	api   map[int]APIPaths
	other GeneralPaths
}

// GeneralPaths holds all non-api route paths
type GeneralPaths struct {
	ConfigurationEndpoint string
	OIDCRedirectEndpoint  string
	JWKSEndpoint          string
	ConsentEndpoint       string
	Privacy               string
}

// APIPaths holds all api route paths
type APIPaths struct {
	MytokenEndpoint       string
	AccessTokenEndpoint   string
	TokenInfoEndpoint     string
	RevocationEndpoint    string
	TokenTransferEndpoint string
	UserSettingEndpoint   string
}

// GetCurrentAPIPaths returns the api paths for the most recent major version
func GetCurrentAPIPaths() APIPaths {
	return GetAPIPaths(version.MAJOR)
}

// GetAPIPaths returns the api paths for the passed major version
func GetAPIPaths(apiVersion int) APIPaths {
	return routes.api[apiVersion]
}

// GetGeneralPaths returns the non-API paths
func GetGeneralPaths() GeneralPaths {
	return routes.other
}
