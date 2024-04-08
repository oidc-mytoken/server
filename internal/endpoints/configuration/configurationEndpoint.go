package configuration

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils"
	gossh "golang.org/x/crypto/ssh"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/endpoints/configuration/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/oidc/oidcfed"
	"github.com/oidc-mytoken/server/internal/server/paths"
)

func SupportedProviders() []api.SupportedProviderConfig {
	if config.Get().Features.Federation.Enabled {
		mytokenConfig.ProvidersSupported = append(getProvidersFromConfig(), oidcfed.SupportedProviders()...)
	} else {
		mytokenConfig.ProvidersSupported = getProvidersFromConfig()
	}
	return mytokenConfig.ProvidersSupported
}

// HandleConfiguration handles calls to the configuration endpoint
func HandleConfiguration(*fiber.Ctx) *model.Response {
	if config.Get().Features.Federation.Enabled {
		mytokenConfig.ProvidersSupported = append(getProvidersFromConfig(), oidcfed.SupportedProviders()...)
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: mytokenConfig,
	}
}

var mytokenConfig *pkg.MytokenConfiguration

var configProviders []api.SupportedProviderConfig

func getProvidersFromConfig() []api.SupportedProviderConfig {
	if configProviders != nil {
		return configProviders
	}
	for _, p := range config.Get().Providers {
		configProviders = append(
			configProviders, api.SupportedProviderConfig{
				Issuer:          p.Issuer,
				Name:            p.Name,
				ScopesSupported: p.Scopes,
			},
		)
	}
	return configProviders
}

// Init initializes the configuration endpoint
func Init() {
	mytokenConfig = basicConfiguration()
	addTokenRevocation(mytokenConfig)
	addShortTokens(mytokenConfig)
	addTransferCodes(mytokenConfig)
	addPollingCodes(mytokenConfig)
	addTokenInfo(mytokenConfig)
	addSSHGrant(mytokenConfig)
}

func basicConfiguration() *pkg.MytokenConfiguration {
	apiPaths := paths.GetCurrentAPIPaths()
	otherPaths := paths.GetGeneralPaths()
	return &pkg.MytokenConfiguration{
		MytokenConfiguration: api.MytokenConfiguration{
			Issuer:                config.Get().IssuerURL,
			AccessTokenEndpoint:   utils.CombineURLPath(config.Get().IssuerURL, apiPaths.AccessTokenEndpoint),
			MytokenEndpoint:       utils.CombineURLPath(config.Get().IssuerURL, apiPaths.MytokenEndpoint),
			TokeninfoEndpoint:     utils.CombineURLPath(config.Get().IssuerURL, apiPaths.TokenInfoEndpoint),
			UserSettingsEndpoint:  utils.CombineURLPath(config.Get().IssuerURL, apiPaths.UserSettingEndpoint),
			NotificationsEndpoint: utils.CombineURLPath(config.Get().IssuerURL, apiPaths.NotificationEndpoint),
			ProfilesEndpoint:      utils.CombineURLPath(config.Get().IssuerURL, apiPaths.ProfilesEndpoint),
			JWKSURI:               utils.CombineURLPath(config.Get().IssuerURL, otherPaths.JWKSEndpoint),
			ProvidersSupported:    getProvidersFromConfig(),
			TokenSigningAlgValue:  config.Get().Signing.Mytoken.Alg.String(),
			ServiceDocumentation:  config.Get().ServiceDocumentation,
			Version:               version.VERSION,
		},
		AccessTokenEndpointGrantTypesSupported: []model.GrantType{model.GrantTypeMytoken},
		MytokenEndpointGrantTypesSupported: []model.GrantType{
			model.GrantTypeOIDCFlow,
			model.GrantTypeMytoken,
		},
		MytokenEndpointOIDCFlowsSupported: []model.OIDCFlow{model.OIDCFlowAuthorizationCode},
		ResponseTypesSupported:            []model.ResponseType{model.ResponseTypeToken},
		TokenEndpoint: utils.CombineURLPath(
			config.Get().IssuerURL, apiPaths.AccessTokenEndpoint,
		),
		RestrictionClaimsSupported: model.AllRestrictionClaims.Disable(config.Get().Features.DisabledRestrictionKeys),
	}
}

func addTokenRevocation(mytokenConfig *pkg.MytokenConfiguration) {
	if config.Get().Features.TokenRevocation.Enabled {
		mytokenConfig.RevocationEndpoint = utils.CombineURLPath(
			config.Get().IssuerURL,
			paths.GetCurrentAPIPaths().RevocationEndpoint,
		)
	}
}
func addShortTokens(mytokenConfig *pkg.MytokenConfiguration) {
	if config.Get().Features.ShortTokens.Enabled {
		model.ResponseTypeShortToken.AddToSliceIfNotFound(&mytokenConfig.ResponseTypesSupported)
	}
}
func addTransferCodes(mytokenConfig *pkg.MytokenConfiguration) {
	if config.Get().Features.TransferCodes.Enabled {
		mytokenConfig.TokenTransferEndpoint = utils.CombineURLPath(
			config.Get().IssuerURL,
			paths.GetCurrentAPIPaths().TokenTransferEndpoint,
		)
		model.GrantTypeTransferCode.AddToSliceIfNotFound(&mytokenConfig.MytokenEndpointGrantTypesSupported)
		model.ResponseTypeTransferCode.AddToSliceIfNotFound(&mytokenConfig.ResponseTypesSupported)
	}
}
func addPollingCodes(mytokenConfig *pkg.MytokenConfiguration) {
	if config.Get().Features.Polling.Enabled {
		model.GrantTypePollingCode.AddToSliceIfNotFound(&mytokenConfig.MytokenEndpointGrantTypesSupported)
	}
}
func addTokenInfo(mytokenConfig *pkg.MytokenConfiguration) {
	if !config.Get().Features.TokenInfo.Enabled {
		mytokenConfig.TokeninfoEndpoint = ""
	} else {
		if config.Get().Features.TokenInfo.Introspect.Enabled {
			model.TokeninfoActionIntrospect.AddToSliceIfNotFound(&mytokenConfig.TokenInfoEndpointActionsSupported)
		}
		if config.Get().Features.TokenInfo.History.Enabled {
			model.TokeninfoActionEventHistory.AddToSliceIfNotFound(&mytokenConfig.TokenInfoEndpointActionsSupported)
		}
		if config.Get().Features.TokenInfo.Tree.Enabled {
			model.TokeninfoActionSubtokenTree.AddToSliceIfNotFound(&mytokenConfig.TokenInfoEndpointActionsSupported)
		}
		if config.Get().Features.TokenInfo.List.Enabled {
			model.TokeninfoActionListMytokens.AddToSliceIfNotFound(&mytokenConfig.TokenInfoEndpointActionsSupported)
		}
	}
}
func addSSHGrant(mytokenconfig *pkg.MytokenConfiguration) {
	if config.Get().Features.SSH.Enabled {
		model.GrantTypeSSH.AddToSliceIfNotFound(&mytokenconfig.MytokenEndpointGrantTypesSupported)
		mytokenconfig.SSHKeys = createSSHKeyInfos()
	}
}

func createSSHKeyInfos() []api.SSHKeyMetadata {
	keys := make([]api.SSHKeyMetadata, len(config.Get().Features.SSH.PrivateKeys))
	for i, sk := range config.Get().Features.SSH.PrivateKeys {
		pk := sk.PublicKey()
		keyType := pk.Type()
		keys[i] = api.SSHKeyMetadata{
			Type:        keyType,
			Fingerprint: gossh.FingerprintSHA256(pk),
		}
	}
	return keys
}
