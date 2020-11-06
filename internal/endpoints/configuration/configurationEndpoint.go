package configuration

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/endpoints/configuration/pkg"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/server/apiPath"
	"github.com/zachmann/mytoken/internal/utils"
)

func HandleConfiguration(ctx *fiber.Ctx) error {
	res := model.Response{
		Status:   fiber.StatusOK,
		Response: mytokenConfig,
	}
	return res.Send(ctx)
}

var mytokenConfig *pkg.MytokenConfiguration

func getProvidersFromConfig() (providers []pkg.SupportedProviderConfig) {
	for _, p := range config.Get().Providers {
		providers = append(providers, pkg.SupportedProviderConfig{
			Issuer:          p.Issuer,
			ScopesSupported: p.Scopes,
		})
	}
	return
}

func Init() {
	mytokenConfig = &pkg.MytokenConfiguration{
		Issuer:                                 config.Get().IssuerURL,
		AccessTokenEndpoint:                    utils.CombineURLPath(config.Get().IssuerURL, apiPath.CURRENT, "/token/access"),
		SuperTokenEndpoint:                     utils.CombineURLPath(config.Get().IssuerURL, apiPath.CURRENT, "/token/super"),
		TokeninfoEndpoint:                      utils.CombineURLPath(config.Get().IssuerURL, apiPath.CURRENT, "/tokeninfo"),
		RevocationEndpoint:                     utils.CombineURLPath(config.Get().IssuerURL, apiPath.CURRENT, "/revocation"),
		JWKSURI:                                utils.CombineURLPath(config.Get().IssuerURL, "/jwks"),
		ProvidersSupported:                     getProvidersFromConfig(),
		TokenSigningAlgValue:                   config.Get().Signing.Alg,
		AccessTokenEndpointGrantTypesSupported: []model.GrantType{model.GrantTypeSuperToken},
		SuperTokenEndpointGrantTypesSupported:  config.Get().EnabledSuperTokenEndpointGrantTypes,
		SuperTokenEndpointOIDCFlowsSupported:   config.Get().EnabledOIDCFlows,
		ServiceDocumentation:                   config.Get().ServiceDocumentation,
	}
}
