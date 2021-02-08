package consent

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/internal/endpoints/consent/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/authcode"
	model2 "github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/supertoken/capabilities"
	"github.com/oidc-mytoken/server/shared/supertoken/restrictions"
)

// handleConsent displays a consent page
func handleConsent(ctx *fiber.Ctx, r restrictions.Restrictions, c capabilities.Capabilities, sc capabilities.Capabilities) error {
	binding := map[string]interface{}{
		"consent":      true,
		"empty-navbar": true,
		"restrictions": pkg.WebRestrictions{Restrictions: r},
		"capabilities": pkg.WebCapabilities(c),
	}
	if c.Has(capabilities.CapabilityCreateST) {
		binding["subtoken-capabilities"] = pkg.WebCapabilities(sc)
	}
	return ctx.Render("sites/consent", binding, "layouts/main")
}

// HandleConsent displays a consent page
func HandleConsent(ctx *fiber.Ctx) error {
	consentCode := state.ParseConsentCode(ctx.Params("consent_code"))
	state := state.NewState(consentCode.GetState())
	authInfo, err := authcodeinforepo.GetAuthFlowInfoByState(state)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.ErrNotFound
		}
		return err
	}
	return handleConsent(ctx, authInfo.Restrictions, authInfo.Capabilities, authInfo.SubtokenCapabilities)
}

func HandleConsentPost(ctx *fiber.Ctx) error {
	consentCode := state.ParseConsentCode(ctx.Params("consent_code"))
	oState := state.NewState(consentCode.GetState())
	authInfo, err := authcodeinforepo.GetAuthFlowInfoByState(oState)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.ErrNotFound
		}
		return err
	}
	authURL := strings.Replace(authInfo.AuthorizationURL, authcode.StatePlaceHolder, oState.State(), 1)
	req := pkg.ConsentPostRequest{}
	if err = json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	for _, c := range req.Capabilities {
		if !capabilities.AllCapabilities.Has(c) {
			return model.Response{
				Status:   fiber.StatusBadRequest,
				Response: model2.BadRequestError(fmt.Sprintf("unknown capability '%s'", c)),
			}.Send(ctx)
		}
	}
	for _, c := range req.SubtokenCapabilities {
		if !capabilities.AllCapabilities.Has(c) {
			return model.Response{
				Status:   fiber.StatusBadRequest,
				Response: model2.BadRequestError(fmt.Sprintf("unknown subtoken_capability '%s'", c)),
			}.Send(ctx)
		}
	}
	if err = authcodeinforepo.UpdateTokenInfoByState(nil, oState, req.Restrictions, req.Capabilities, req.SubtokenCapabilities); err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	return model.Response{
		Status: 278,
		Response: map[string]string{
			"authorization_url": authURL,
		},
	}.Send(ctx)
}
