package configuration

import (
	"github.com/gofiber/fiber/v2"

	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/endpoints/configuration/pkg"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/server/apiPath"
	"github.com/zachmann/mytoken/internal/utils"
)

// HandleConfiguration handles calls to the configuration endpoint
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

// Init initializes the configuration endpoint
func Init() {
	mytokenConfig = &pkg.MytokenConfiguration{
		Issuer:                                 config.Get().IssuerURL,
		AccessTokenEndpoint:                    utils.CombineURLPath(config.Get().IssuerURL, apiPath.CURRENT, "/token/access"),
		SuperTokenEndpoint:                     utils.CombineURLPath(config.Get().IssuerURL, apiPath.CURRENT, "/token/super"),
		TokeninfoEndpoint:                      utils.CombineURLPath(config.Get().IssuerURL, apiPath.CURRENT, "/tokeninfo"),
		UserSettingsEndpoint:                   utils.CombineURLPath(config.Get().IssuerURL, apiPath.CURRENT, "/user"),
		JWKSURI:                                utils.CombineURLPath(config.Get().IssuerURL, "/jwks"),
		ProvidersSupported:                     getProvidersFromConfig(),
		TokenSigningAlgValue:                   config.Get().Signing.Alg,
		AccessTokenEndpointGrantTypesSupported: []model.GrantType{model.GrantTypeSuperToken},
		SuperTokenEndpointGrantTypesSupported:  []model.GrantType{model.GrantTypeOIDCFlow, model.GrantTypeSuperToken},
		SuperTokenEndpointOIDCFlowsSupported:   config.Get().Features.EnabledOIDCFlows,
		ResponseTypesSupported:                 []model.ResponseType{model.ResponseTypeToken},
		ServiceDocumentation:                   config.Get().ServiceDocumentation,
		Version:                                model.VERSION,
	}
	if config.Get().Features.TokenRevocation.Enabled {
		mytokenConfig.RevocationEndpoint = utils.CombineURLPath(config.Get().IssuerURL, apiPath.CURRENT, "/token/revoke")
	}
	if config.Get().Features.TransferCodes.Enabled {
		model.ResponseTypeShortToken.AddToSliceIfNotFound(&mytokenConfig.ResponseTypesSupported)
	}
	if config.Get().Features.TransferCodes.Enabled {
		mytokenConfig.TokenTransferEndpoint = utils.CombineURLPath(config.Get().IssuerURL, apiPath.CURRENT, "/token/transfer")
		model.GrantTypeTransferCode.AddToSliceIfNotFound(&mytokenConfig.SuperTokenEndpointGrantTypesSupported)
		model.ResponseTypeTransferCode.AddToSliceIfNotFound(&mytokenConfig.ResponseTypesSupported)
	}
	if config.Get().Features.Polling.Enabled {
		model.GrantTypePollingCode.AddToSliceIfNotFound(&mytokenConfig.SuperTokenEndpointGrantTypesSupported)
	}
	if config.Get().Features.AccessTokenGrant.Enabled {
		model.GrantTypeAccessToken.AddToSliceIfNotFound(&mytokenConfig.SuperTokenEndpointGrantTypesSupported)
	}
	if config.Get().Features.SignedJWTGrant.Enabled {
		model.GrantTypePrivateKeyJWT.AddToSliceIfNotFound(&mytokenConfig.SuperTokenEndpointGrantTypesSupported)
	}
}
