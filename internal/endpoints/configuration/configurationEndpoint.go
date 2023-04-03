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
	"github.com/oidc-mytoken/server/internal/server/routes"
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
		providers = append(
			providers, api.SupportedProviderConfig{
				Issuer:          p.Issuer,
				Name:            p.Name,
				ScopesSupported: p.Scopes,
			},
		)
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
	addTokenInfo(mytokenConfig)
	addSSHGrant(mytokenConfig)
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
			ProfilesEndpoint:     utils.CombineURLPath(config.Get().IssuerURL, apiPaths.ProfilesEndpoint),
			JWKSURI:              utils.CombineURLPath(config.Get().IssuerURL, otherPaths.JWKSEndpoint),
			ProvidersSupported:   getProvidersFromConfig(),
			TokenSigningAlgValue: config.Get().Signing.Alg.String(),
			ServiceDocumentation: config.Get().ServiceDocumentation,
			Version:              version.VERSION,
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
			routes.GetCurrentAPIPaths().RevocationEndpoint,
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
			routes.GetCurrentAPIPaths().TokenTransferEndpoint,
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
