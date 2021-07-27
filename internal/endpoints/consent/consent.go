package consent

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/internal/endpoints/consent/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/authcode"
	model2 "github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/utils"
)

// handleConsent displays a consent page
func handleConsent(ctx *fiber.Ctx, authInfo *authcodeinforepo.AuthFlowInfoOut) error {
	c := authInfo.Capabilities
	sc := authInfo.SubtokenCapabilities
	binding := map[string]interface{}{
		"consent":          true,
		"empty-navbar":     true,
		"restr-gui":        true,
		"restrictions":     pkg.WebRestrictions{Restrictions: authInfo.Restrictions},
		"capabilities":     pkg.WebCapabilities(c),
		"iss":              authInfo.Issuer,
		"supported_scopes": strings.Join(config.Get().ProviderByIssuer[authInfo.Issuer].Scopes, " "),
		"token-name":       authInfo.Name,
		"rotation":         authInfo.Rotation,
	}
	if c.Has(api.CapabilityCreateMT) {
		if len(sc) == 0 {
			sc = c
		}
		binding["subtoken-capabilities"] = pkg.WebCapabilities(sc)
	}
	return ctx.Render("sites/consent", binding, "layouts/main-no-container")
}

func getAuthInfoFromConsentCodeStr(code string) (*authcodeinforepo.AuthFlowInfoOut, *state.State, error) {
	consentCode := state.ConsentCodeFromStr(code)
	oState := state.NewState(consentCode.GetState())
	authInfo, err := authcodeinforepo.GetAuthFlowInfoByState(oState)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fiber.ErrNotFound
		}
	}
	return authInfo, oState, err
}

// HandleConsent displays a consent page
func HandleConsent(ctx *fiber.Ctx) error {
	authInfo, _, err := getAuthInfoFromConsentCodeStr(ctx.Params("consent_code"))
	if err != nil {
		return err
	}
	return handleConsent(ctx, authInfo)
}

func handleConsentDecline(ctx *fiber.Ctx, authInfo *authcodeinforepo.AuthFlowInfoOut, oState *state.State) error {
	url := "/"
	if authInfo.PollingCode {
		url = "/native/abort"
	}
	if err := authcodeinforepo.DeleteAuthFlowInfoByState(nil, oState); err != nil {
		res := model.ErrorToInternalServerErrorResponse(err)
		m := utils.StructToStringMapUsingJSONTags(res.Response)
		m["url"] = url
		res.Response = m
		return res.Send(ctx)
	}
	if authInfo.PollingCode {
		if err := transfercoderepo.DeclineConsentByState(nil, oState); err != nil {
			log.WithError(err).Error()
		}
	}
	return model.Response{
		Status: 278,
		Response: map[string]string{
			"url": url,
		},
	}.Send(ctx)
}

// HandleConsentPost handles consent confirmation requests
func HandleConsentPost(ctx *fiber.Ctx) error {
	authInfo, oState, err := getAuthInfoFromConsentCodeStr(ctx.Params("consent_code"))
	if err != nil {
		return err
	}
	if len(ctx.Body()) == 0 {
		return handleConsentDecline(ctx, authInfo, oState)
	}
	req := pkg.ConsentPostRequest{}
	if err = json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	for _, c := range req.Capabilities {
		if !api.AllCapabilities.Has(c) {
			return model.Response{
				Status:   fiber.StatusBadRequest,
				Response: model2.BadRequestError(fmt.Sprintf("unknown capability '%s'", c)),
			}.Send(ctx)
		}
	}
	for _, c := range req.SubtokenCapabilities {
		if !api.AllCapabilities.Has(c) {
			return model.Response{
				Status:   fiber.StatusBadRequest,
				Response: model2.BadRequestError(fmt.Sprintf("unknown subtoken_capability '%s'", c)),
			}.Send(ctx)
		}
	}
	if err = authcodeinforepo.UpdateTokenInfoByState(
		nil, oState, req.Restrictions, req.Capabilities, req.SubtokenCapabilities, req.Rotation, req.TokenName,
	); err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	provider, ok := config.Get().ProviderByIssuer[req.Issuer]
	if !ok {
		return model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}.Send(ctx)
	}
	authURL := authcode.GetAuthorizationURL(provider, oState.State(), req.Restrictions)
	return model.Response{
		Status: 278,
		Response: map[string]string{
			"authorization_url": authURL,
		},
	}.Send(ctx)
}
