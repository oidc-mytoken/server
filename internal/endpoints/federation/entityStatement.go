package federation

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/utils/utils"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	oidcfed "github.com/zachmann/go-oidcfed/pkg"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/jws"
	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/server/routes"
)

var entityConfiguration *oidcfed.EntityConfiguration
var entityConfigurationRes entityStatementResponse

func getEntityConfigurationRes() entityStatementResponse {
	if entityConfiguration == nil {
		initEntityConfiguration()
	}
	if entityConfiguration.ExpiresAt <= time.Now().Unix() {
		initEntityConfiguration()
	}
	return entityConfigurationRes
}

func initEntityConfiguration() {
	otherPaths := routes.GetGeneralPaths()
	privacyURI := utils.CombineURLPath(config.Get().IssuerURL, otherPaths.Privacy)
	entityConfiguration = oidcfed.NewEntityConfiguration(
		oidcfed.EntityStatementPayload{
			Issuer:         config.Get().IssuerURL,
			Subject:        config.Get().IssuerURL,
			IssuedAt:       time.Now().Unix(),
			ExpiresAt:      time.Now().Add(time.Duration(config.Get().Features.Federation.EntityConfigurationLifetime) * time.Second).Unix(),
			JWKS:           jws.GetJWKS(jws.KeyUsageFederation),
			AuthorityHints: config.Get().Features.Federation.AuthorityHints,
			Metadata: &oidcfed.Metadata{
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
					JWKSURI:                 utils.CombineURLPath(config.Get().IssuerURL, otherPaths.JWKSEndpoint),
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
		}, jws.GetSigningKey(jws.KeyUsageFederation), config.Get().Features.Federation.Signing.Alg,
	)
	entityConfigurationJWT, err := entityConfiguration.JWT()
	if err != nil {
		log.WithError(err).Fatal("could not create entity configuration JWT")
	}
	entityConfigurationRes = entityConfigurationJWT
}

func Init() {
	if !config.Get().Features.Federation.Enabled {
		return
	}
	jws.LoadFederationKey()
	initEntityConfiguration()
}

type entityStatementResponse []byte

// Send sends this response using the passed fiber.Ctx
func (r entityStatementResponse) Send(ctx *fiber.Ctx) error {
	ctx.Set("content-type", "application/entity-statement+jwt")
	return ctx.Status(fasthttp.StatusOK).Send(r)
}

// HandleEntityConfiguration handles calls to the oidc federation entity configuration endpoint
func HandleEntityConfiguration(ctx *fiber.Ctx) error {
	return getEntityConfigurationRes().Send(ctx)
}
