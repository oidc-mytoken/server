package federation

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	oidcfed "github.com/zachmann/go-oidcfed/pkg"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/jws"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/server/routes"
)

func InitEntityConfiguration() {
	if config.Get().Features.Federation.Entity != nil {
		return
	}
	otherPaths := routes.GetGeneralPaths()
	privacyURI := utils.CombineURLPath(config.Get().IssuerURL, otherPaths.Privacy)
	var err error
	config.Get().Features.Federation.Entity, err = oidcfed.NewFederationLeaf(
		config.Get().IssuerURL,
		config.Get().Features.Federation.AuthorityHints,
		config.Get().Features.Federation.TrustAnchors,
		&oidcfed.Metadata{
			RelyingParty: &oidcfed.OpenIDRelyingPartyMetadata{
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
				JWKS:                    jws.GetJWKS(jws.KeyUsageOIDCSigning),
				SoftwareID:              version.SOFTWAREID,
				SoftwareVersion:         version.VERSION,
				OrganizationName:        config.Get().ServiceOperator.Name,
				ClientRegistrationTypes: []string{oidcfed.ClientRegistrationTypeAutomatic},
			},
			FederationEntity: &oidcfed.FederationEntityMetadata{
				OrganizationName: config.Get().ServiceOperator.Name,
				Contacts:         []string{config.Get().ServiceOperator.Contact},
				LogoURI:          utils.CombineURLPath(config.Get().IssuerURL, "static/img/mytoken.png"),
				PolicyURI:        privacyURI,
				HomepageURI:      "https://mytoken-docs.data.kit.edu",
			},
		},
		jws.GetSigningKey(jws.KeyUsageFederation),
		config.Get().Features.Federation.Signing.Alg,
		config.Get().Features.Federation.EntityConfigurationLifetime,
	)
	if err != nil {
		log.WithError(err).Fatal("Could not create oidcfed leaf entity configuration")
	}
}

type entityStatementResponse []byte

// Send sends this response using the passed fiber.Ctx
func (r entityStatementResponse) Send(ctx *fiber.Ctx) error {
	ctx.Set("content-type", "application/entity-statement+jwt")
	return ctx.Status(fasthttp.StatusOK).Send(r)
}

// HandleEntityConfiguration handles calls to the oidc federation entity configuration endpoint
func HandleEntityConfiguration(ctx *fiber.Ctx) error {
	entityConfigurationJWT, err := config.Get().Features.Federation.Entity.EntityConfiguration().JWT()
	if err != nil {
		err = errors.Wrap(err, "could not create entity configuration JWT")
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return entityStatementResponse(entityConfigurationJWT).Send(ctx)
}
