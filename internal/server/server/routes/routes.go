package routes

import (
	"github.com/zachmann/mytoken/internal/server/model/version"
	"github.com/zachmann/mytoken/internal/server/server/apiPath"
	"github.com/zachmann/mytoken/internal/utils"
)

var routes *paths

// init initializes the server route paths
func init() {
	routes = &paths{
		api: map[int]APIPaths{
			0: {
				SuperTokenEndpoint:    utils.CombineURLPath(apiPath.V0, "/token/super"),
				AccessTokenEndpoint:   utils.CombineURLPath(apiPath.V0, "/token/access"),
				TokenInfoEndpoint:     utils.CombineURLPath(apiPath.V0, "/tokeninfo"),
				RevocationEndpoint:    utils.CombineURLPath(apiPath.V0, "/token/revoke"),
				TokenTransferEndpoint: utils.CombineURLPath(apiPath.V0, "/token/transfer"),
				UserSettingEndpoint:   utils.CombineURLPath(apiPath.V0, "/user"),
			},
		},
		other: GeneralPaths{
			ConfigurationEndpoint:  "/.well-known/mytoken-configuration",
			OIDCRedirectEndpoint:   "/redirect",
			JWKSEndpoint:           "/jwks",
			NativeRedirectEndpoint: "/n",
		},
	}
}

type paths struct {
	api   map[int]APIPaths
	other GeneralPaths
}

// GeneralPaths holds all non-api route paths
type GeneralPaths struct {
	ConfigurationEndpoint  string
	OIDCRedirectEndpoint   string
	JWKSEndpoint           string
	NativeRedirectEndpoint string
}

// APIPaths holds all api route paths
type APIPaths struct {
	SuperTokenEndpoint    string
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
