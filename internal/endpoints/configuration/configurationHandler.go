package configuration

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/utils"
)

func HandleConfiguration(ctx *fiber.Ctx) error {
	return ctx.JSON(mytokenConfig)
}

var mytokenConfig *MytokenConfiguration

func getProvidersFromConfig() (providers []SupportedProviderConfig) {
	for _, p := range config.Get().Providers {
		providers=append(providers,SupportedProviderConfig{
			Issuer: p.Issuer,
			ScopesSupported: p.Scopes,
		})
	}
	return
}

var apiPath = "/api/v0"

func Init() {
	mytokenConfig = &MytokenConfiguration{
		Issuer: config.Get().IssuerURL,
		AccessTokenEndpoint: utils.CombineURLPath(config.Get().IssuerURL, apiPath, "/token/access"),
		SuperTokenEndpoint: utils.CombineURLPath(config.Get().IssuerURL, apiPath, "/token/super"),
		TokeninfoEndpoint: utils.CombineURLPath(config.Get().IssuerURL, apiPath, "/tokeninfo"),
		RevocationEndpoint: utils.CombineURLPath(config.Get().IssuerURL, apiPath, "/revocation"),
		JWKSURI: utils.CombineURLPath(config.Get().IssuerURL, "/jwks"),
		ProvidersSupported: getProvidersFromConfig(),
		TokenSigningAlgValue: config.Get().TokenSigningAlg,
		AccessTokenEndpointGrantTypesSupported: []model.GrantType{model.GrantTypeSuperToken},
		SuperTokenEndpointGrantTypesSupported: config.Get().EnabledSuperTokenEndpointGrantTypes,
		SuperTokenEndpointOIDCFlowsSupported: config.Get().EnabledOIDCFlows,
		ServiceDocumentation: config.Get().ServiceDocumentation,
	}
}