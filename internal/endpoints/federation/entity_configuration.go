package federation

import (
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

// InitEntityConfiguration initializes the entity configuration if enabled
func InitEntityConfiguration() {
	if config.Get().Features.Federation.Entity != nil {
		return
	}
	otherPaths := paths.GetGeneralPaths()
	privacyURI := utils.CombineURLPath(config.Get().IssuerURL, otherPaths.Privacy)
	var err error
	jwks := jws.GetJWKS(jws.KeyUsageOIDCSigning)
	config.Get().Features.Federation.Entity, err = oidfed.NewFederationLeaf(
		config.Get().IssuerURL,
		config.Get().Features.Federation.AuthorityHints,
		config.Get().Features.Federation.TrustAnchors,
		&oidfed.Metadata{
			RelyingParty: &oidfed.OpenIDRelyingPartyMetadata{
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
