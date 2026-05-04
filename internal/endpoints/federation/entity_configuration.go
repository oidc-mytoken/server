package federation

import (
	"strings"
	"time"

	oidfed "github.com/go-oidfed/lib"
	"github.com/go-oidfed/lib/jwx"
	"github.com/go-oidfed/lib/oidfedconst"
	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/jws"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/server/paths"
)

// defaultOIDCScopes contains the default OIDC core scopes
const defaultOIDCScopes = "openid profile email address phone offline_access"

// InitEntityConfiguration initializes the entity configuration if enabled.
func InitEntityConfiguration() {
	if config.Get().Features.Federation.Entity != nil {
		return
	}
	otherPaths := paths.GetGeneralPaths()
	privacyURI := utils.CombineURLPath(config.Get().IssuerURL, otherPaths.Privacy)
	var err error
	jwks := jws.GetJWKS(jws.KeyUsageOIDCSigning)

	// Use configured scopes if set, otherwise use default OIDC core scopes
	scope := defaultOIDCScopes
	if len(config.Get().Features.Federation.OPDiscovery.Scopes) > 0 {
		scope = strings.Join(config.Get().Features.Federation.OPDiscovery.Scopes, " ")
	}

	config.Get().Features.Federation.Entity, err = oidfed.NewFederationLeaf(
		config.Get().IssuerURL,
		config.Get().Features.Federation.AuthorityHints,
		config.Get().Features.Federation.TrustAnchors,
		&oidfed.Metadata{
			RelyingParty: &oidfed.OpenIDRelyingPartyMetadata{
				Scope: scope,
				RedirectURIS: []string{
					utils.CombineURLPath(
						config.Get().IssuerURL, otherPaths.OIDCRedirectEndpoint,
					),
				},
				GrantTypes: []string{
					"refresh_token",
					"authorization_code",
				},
				ApplicationType:         "web",
				Contacts:                []string{config.Get().ServiceOperator.Contact},
				ClientName:              "mytoken",
				LogoURI:                 utils.CombineURLPath(config.Get().IssuerURL, "static/img/mytoken.png"),
				ClientURI:               config.Get().IssuerURL,
				PolicyURI:               privacyURI,
				TOSURI:                  privacyURI,
				JWKS:                    &jwks,
				SoftwareID:              version.SOFTWAREID,
				SoftwareVersion:         version.VERSION,
				OrganizationName:        config.Get().ServiceOperator.Name,
				ClientRegistrationTypes: []string{oidfedconst.ClientRegistrationTypeAutomatic},
				InformationURI:          "https://docs.mytok.eu",
			},
		},
		jwx.NewEntityStatementSigner(
			jws.GetVersatileSigner(jws.KeyUsageFederation),
		),
		time.Duration(config.Get().Features.Federation.EntityConfigurationLifetime)*time.Second,
		jws.GetVersatileSigner(jws.KeyUsageOIDCSigning),
		nil,
	)
	if err != nil {
		log.WithError(err).Fatal("Could not create oidfed leaf entity configuration")
	}
}

// UpdateScopes updates the scope in the RP metadata of the entity configuration.
// If static scopes are configured in the config file, this function does nothing.
func UpdateScopes(scopes []string) {
	// Skip dynamic update if static scopes are configured
	if len(config.Get().Features.Federation.OPDiscovery.Scopes) > 0 {
		return
	}

	entity := config.Get().Features.Federation.Entity
	if entity == nil {
		return
	}
	// FederationLeaf embeds StaticFederationEntity as a value (not pointer)
	staticEntity, ok := entity.FederationEntity.(oidfed.StaticFederationEntity)
	if !ok {
		return
	}
	if staticEntity.Metadata != nil && staticEntity.Metadata.RelyingParty != nil {
		staticEntity.Metadata.RelyingParty.Scope = strings.Join(scopes, " ")
	}
}

type entityStatementResponse []byte

// Send sends this response using the passed fiber.Ctx
func (r entityStatementResponse) Send(ctx *fiber.Ctx) error {
	ctx.Set("content-type", oidfedconst.ContentTypeEntityStatement)
	return ctx.Status(fasthttp.StatusOK).Send(r)
}

// HandleEntityConfiguration handles calls to the oidc federation entity configuration endpoint
func HandleEntityConfiguration(ctx *fiber.Ctx) error {
	entityConfigurationJWT, err := config.Get().Features.Federation.Entity.EntityConfigurationJWT()
	if err != nil {
		err = errors.Wrap(err, "could not create entity configuration JWT")
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return entityStatementResponse(entityConfigurationJWT).Send(ctx)
}
