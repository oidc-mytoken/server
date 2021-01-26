package configuration

import (
	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/server/config"

	"github.com/oidc-mytoken/server/internal/endpoints/configuration/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/server/routes"
	pkgModel "github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/utils"
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
	apiPaths := routes.GetCurrentAPIPaths()
	otherPaths := routes.GetGeneralPaths()
	mytokenConfig = &pkg.MytokenConfiguration{
		Issuer:                                 config.Get().IssuerURL,
		AccessTokenEndpoint:                    utils.CombineURLPath(config.Get().IssuerURL, apiPaths.AccessTokenEndpoint),
		SuperTokenEndpoint:                     utils.CombineURLPath(config.Get().IssuerURL, apiPaths.SuperTokenEndpoint),
		TokeninfoEndpoint:                      utils.CombineURLPath(config.Get().IssuerURL, apiPaths.TokenInfoEndpoint),
		UserSettingsEndpoint:                   utils.CombineURLPath(config.Get().IssuerURL, apiPaths.UserSettingEndpoint),
		JWKSURI:                                utils.CombineURLPath(config.Get().IssuerURL, otherPaths.JWKSEndpoint),
		ProvidersSupported:                     getProvidersFromConfig(),
		TokenSigningAlgValue:                   config.Get().Signing.Alg,
		AccessTokenEndpointGrantTypesSupported: []pkgModel.GrantType{pkgModel.GrantTypeSuperToken},
		SuperTokenEndpointGrantTypesSupported:  []pkgModel.GrantType{pkgModel.GrantTypeOIDCFlow, pkgModel.GrantTypeSuperToken},
		SuperTokenEndpointOIDCFlowsSupported:   config.Get().Features.EnabledOIDCFlows,
		ResponseTypesSupported:                 []pkgModel.ResponseType{pkgModel.ResponseTypeToken},
		ServiceDocumentation:                   config.Get().ServiceDocumentation,
		Version:                                version.VERSION,
	}
	if config.Get().Features.TokenRevocation.Enabled {
		mytokenConfig.RevocationEndpoint = utils.CombineURLPath(config.Get().IssuerURL, apiPaths.RevocationEndpoint)
	}
	if config.Get().Features.TransferCodes.Enabled {
		pkgModel.ResponseTypeShortToken.AddToSliceIfNotFound(&mytokenConfig.ResponseTypesSupported)
	}
	if config.Get().Features.TransferCodes.Enabled {
		mytokenConfig.TokenTransferEndpoint = utils.CombineURLPath(config.Get().IssuerURL, apiPaths.TokenTransferEndpoint)
		pkgModel.GrantTypeTransferCode.AddToSliceIfNotFound(&mytokenConfig.SuperTokenEndpointGrantTypesSupported)
		pkgModel.ResponseTypeTransferCode.AddToSliceIfNotFound(&mytokenConfig.ResponseTypesSupported)
	}
	if config.Get().Features.Polling.Enabled {
		pkgModel.GrantTypePollingCode.AddToSliceIfNotFound(&mytokenConfig.SuperTokenEndpointGrantTypesSupported)
	}
	if config.Get().Features.AccessTokenGrant.Enabled {
		pkgModel.GrantTypeAccessToken.AddToSliceIfNotFound(&mytokenConfig.SuperTokenEndpointGrantTypesSupported)
	}
	if config.Get().Features.SignedJWTGrant.Enabled {
		pkgModel.GrantTypePrivateKeyJWT.AddToSliceIfNotFound(&mytokenConfig.SuperTokenEndpointGrantTypesSupported)
	}
}
