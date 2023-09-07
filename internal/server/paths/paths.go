package paths

import (
	"github.com/oidc-mytoken/utils/utils"

	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/server/apipath"
)

var routes *paths

// WellknownMytokenConfiguration is the mytoken configuration path suffix
const WellknownMytokenConfiguration = "/.well-known/mytoken-configuration"

// WellknownOpenIDConfiguration is the openid configuration path suffix
const WellknownOpenIDConfiguration = "/.well-known/openid-configuration"

// WellknownOpenIDFederation is the openid federation path suffix
const WellknownOpenIDFederation = "/.well-known/openid-federation"

func defaultAPIPaths(api string) APIPaths {
	return APIPaths{
		MytokenEndpoint:       utils.CombineURLPath(api, "/token/my"),
		AccessTokenEndpoint:   utils.CombineURLPath(api, "/token/access"),
		TokenInfoEndpoint:     utils.CombineURLPath(api, "/tokeninfo"),
		RevocationEndpoint:    utils.CombineURLPath(api, "/token/revoke"),
		TokenTransferEndpoint: utils.CombineURLPath(api, "/token/transfer"),
		UserSettingEndpoint:   utils.CombineURLPath(api, "/settings"),
		ProfilesEndpoint:      utils.CombineURLPath(api, "/pt"),
		GuestModeOP:           utils.CombineURLPath(api, "/guests"),
	}
}

// init initializes the server route paths
func init() {
	routes = &paths{
		api: map[int]APIPaths{
			0: defaultAPIPaths(apipath.V0),
		},
		other: GeneralPaths{
			ConfigurationEndpoint: WellknownMytokenConfiguration,
			FederationEndpoint:    WellknownOpenIDFederation,
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
	FederationEndpoint    string
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
	ProfilesEndpoint      string
	GuestModeOP           string
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
