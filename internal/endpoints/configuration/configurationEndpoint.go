package configuration

import (
	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/endpoints/configuration/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/server/routes"
	pkgModel "github.com/oidc-mytoken/server/shared/model"
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

func getProvidersFromConfig() (providers []api.SupportedProviderConfig) {
	for _, p := range config.Get().Providers {
		providers = append(providers, api.SupportedProviderConfig{
			Issuer:          p.Issuer,
			ScopesSupported: p.Scopes,
		})
	}
	return
}

// Init initializes the configuration endpoint
func Init() {
	mytokenConfig = basicConfiguration()
	addTokenRevocation(mytokenConfig)
	addShortTokens(mytokenConfig)
	addTransferCodes(mytokenConfig)
	addPollingCodes(mytokenConfig)
	addAccessTokenGrant(mytokenConfig)
	addSignedJWTGrant(mytokenConfig)
	addTokenInfo(mytokenConfig)
}

func basicConfiguration() *pkg.MytokenConfiguration {
	apiPaths := routes.GetCurrentAPIPaths()
	otherPaths := routes.GetGeneralPaths()
	return &pkg.MytokenConfiguration{
		MytokenConfiguration: api.MytokenConfiguration{
			Issuer:               config.Get().IssuerURL,
			AccessTokenEndpoint:  utils.CombineURLPath(config.Get().IssuerURL, apiPaths.AccessTokenEndpoint),
			MytokenEndpoint:      utils.CombineURLPath(config.Get().IssuerURL, apiPaths.MytokenEndpoint),
			TokeninfoEndpoint:    utils.CombineURLPath(config.Get().IssuerURL, apiPaths.TokenInfoEndpoint),
			UserSettingsEndpoint: utils.CombineURLPath(config.Get().IssuerURL, apiPaths.UserSettingEndpoint),
			JWKSURI:              utils.CombineURLPath(config.Get().IssuerURL, otherPaths.JWKSEndpoint),
			ProvidersSupported:   getProvidersFromConfig(),
			TokenSigningAlgValue: config.Get().Signing.Alg,
			ServiceDocumentation: config.Get().ServiceDocumentation,
			Version:              version.VERSION(),
		},
		AccessTokenEndpointGrantTypesSupported: []pkgModel.GrantType{pkgModel.GrantTypeMytoken},
		MytokenEndpointGrantTypesSupported:     []pkgModel.GrantType{pkgModel.GrantTypeOIDCFlow, pkgModel.GrantTypeMytoken},
		MytokenEndpointOIDCFlowsSupported:      config.Get().Features.EnabledOIDCFlows,
		ResponseTypesSupported:                 []pkgModel.ResponseType{pkgModel.ResponseTypeToken},
		TokenEndpoint:                          utils.CombineURLPath(config.Get().IssuerURL, apiPaths.AccessTokenEndpoint),
		SupportedRestrictionKeys:               model.AllRestrictionKeys.Disable(config.Get().Features.DisabledRestrictionKeys),
	}
}

func addTokenRevocation(mytokenConfig *pkg.MytokenConfiguration) {
	if config.Get().Features.TokenRevocation.Enabled {
		mytokenConfig.RevocationEndpoint = utils.CombineURLPath(config.Get().IssuerURL,
			routes.GetCurrentAPIPaths().RevocationEndpoint)
	}
}
func addShortTokens(mytokenConfig *pkg.MytokenConfiguration) {
	if config.Get().Features.ShortTokens.Enabled {
		pkgModel.ResponseTypeShortToken.AddToSliceIfNotFound(&mytokenConfig.ResponseTypesSupported)
	}
}
func addTransferCodes(mytokenConfig *pkg.MytokenConfiguration) {
	if config.Get().Features.TransferCodes.Enabled {
		mytokenConfig.TokenTransferEndpoint = utils.CombineURLPath(config.Get().IssuerURL,
			routes.GetCurrentAPIPaths().TokenTransferEndpoint)
		pkgModel.GrantTypeTransferCode.AddToSliceIfNotFound(&mytokenConfig.MytokenEndpointGrantTypesSupported)
		pkgModel.ResponseTypeTransferCode.AddToSliceIfNotFound(&mytokenConfig.ResponseTypesSupported)
	}
}
func addPollingCodes(mytokenConfig *pkg.MytokenConfiguration) {
	if config.Get().Features.Polling.Enabled {
		pkgModel.GrantTypePollingCode.AddToSliceIfNotFound(&mytokenConfig.MytokenEndpointGrantTypesSupported)
	}
}
func addAccessTokenGrant(mytokenConfig *pkg.MytokenConfiguration) {
	if config.Get().Features.AccessTokenGrant.Enabled {
		pkgModel.GrantTypeAccessToken.AddToSliceIfNotFound(&mytokenConfig.MytokenEndpointGrantTypesSupported)
	}
}
func addSignedJWTGrant(mytokenConfig *pkg.MytokenConfiguration) {
	if config.Get().Features.SignedJWTGrant.Enabled {
		pkgModel.GrantTypePrivateKeyJWT.AddToSliceIfNotFound(&mytokenConfig.MytokenEndpointGrantTypesSupported)
	}
}
func addTokenInfo(mytokenConfig *pkg.MytokenConfiguration) {
	if !config.Get().Features.TokenInfo.Enabled {
		mytokenConfig.TokeninfoEndpoint = ""
	} else {
		if config.Get().Features.TokenInfo.Introspect.Enabled {
			pkgModel.TokeninfoActionIntrospect.AddToSliceIfNotFound(&mytokenConfig.TokenInfoEndpointActionsSupported)
		}
		if config.Get().Features.TokenInfo.History.Enabled {
			pkgModel.TokeninfoActionEventHistory.AddToSliceIfNotFound(&mytokenConfig.TokenInfoEndpointActionsSupported)
		}
		if config.Get().Features.TokenInfo.Tree.Enabled {
			pkgModel.TokeninfoActionSubtokenTree.AddToSliceIfNotFound(&mytokenConfig.TokenInfoEndpointActionsSupported)
		}
		if config.Get().Features.TokenInfo.List.Enabled {
			pkgModel.TokeninfoActionListMytokens.AddToSliceIfNotFound(&mytokenConfig.TokenInfoEndpointActionsSupported)
		}
	}
}
