package consent

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	pkg2 "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/server/httpStatus"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/internal/utils/templating"

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
func handleConsent(ctx *fiber.Ctx, info *pkg2.OIDCFlowRequest, includeConsentCallbacks bool) error {
	c := info.Capabilities
	binding := map[string]interface{}{
		"consent":               true,
		"consent-send":          includeConsentCallbacks,
		"empty-navbar":          true,
		"restr-gui":             true,
		"collapse":              templating.Collapsable{All: true},
		"restrictions":          pkg.WebRestrictions{Restrictions: info.Restrictions},
		"capabilities":          pkg.AllWebCapabilities(),
		"subtoken-capabilities": pkg.AllWebCapabilities(),
		"checked-capabilities":  c.Strings(),
		"iss":                   info.Issuer,
		"supported_scopes":      strings.Join(config.Get().ProviderByIssuer[info.Issuer].Scopes, " "),
		"token-name":            info.Name,
		"rotation":              info.Rotation,
		"application":           info.ApplicationName,
	}
	if c.Has(api.CapabilityCreateMT) {
		binding["checked-subtoken-capabilities"] = info.SubtokenCapabilities.Strings()
	}
	if !includeConsentCallbacks {
		binding["instance-url"] = config.Get().IssuerURL
	}
	return ctx.Render("sites/consent", binding, "layouts/main")
}

func getAuthInfoFromConsentCodeStr(rlog log.Ext1FieldLogger, code string) (
	*authcodeinforepo.AuthFlowInfoOut, *state.State, error,
) {
	consentCode := state.ConsentCodeFromStr(code)
	oState := state.NewState(consentCode.GetState())
	authInfo, err := authcodeinforepo.GetAuthFlowInfoByState(rlog, oState)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fiber.ErrNotFound
		} else {
			rlog.Errorf("%s", errorfmt.Full(err))
		}
	}
	return authInfo, oState, err
}

// HandleCreateConsent returns a consent page for the posted parameters
func HandleCreateConsent(ctx *fiber.Ctx) error {
	req := pkg.ConsentRequest{}
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	if req.Issuer == "" {
		return model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model2.BadRequestError("required parameter 'oidc_issuer' missing"),
		}.Send(ctx)
	}
	info := &pkg2.OIDCFlowRequest{
		OIDCFlowRequest: api.OIDCFlowRequest{
			GeneralMytokenRequest: api.GeneralMytokenRequest{
				Issuer:               req.Issuer,
				Capabilities:         req.Capabilities,
				SubtokenCapabilities: req.SubtokenCapabilities,
				Name:                 req.TokenName,
				Rotation:             req.Rotation,
				ApplicationName:      req.ApplicationName,
			},
		},
		Restrictions: req.Restrictions,
	}
	return handleConsent(ctx, info, false)
}

// HandleConsent displays a consent page
func HandleConsent(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	authInfo, _, err := getAuthInfoFromConsentCodeStr(rlog, ctx.Params("consent_code"))
	if err != nil {
		// Don't log error here, it was already logged
		return err
	}
	return handleConsent(ctx, &(authInfo.AuthCodeFlowRequest.OIDCFlowRequest), true)
}

func handleConsentDecline(ctx *fiber.Ctx, authInfo *authcodeinforepo.AuthFlowInfoOut, oState *state.State) error {
	rlog := logger.GetRequestLogger(ctx)
	url := "/"
	if authInfo.PollingCode {
		url = "/native/abort"
		if authInfo.ApplicationName != "" {
			url = fmt.Sprintf("%s?application=%s", url, authInfo.ApplicationName)
		}
	}
	if err := authcodeinforepo.DeleteAuthFlowInfoByState(rlog, nil, oState); err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		res := model.ErrorToInternalServerErrorResponse(err)
		m := utils.StructToStringMapUsingJSONTags(res.Response)
		m["url"] = url
		res.Response = m
		return res.Send(ctx)
	}
	if authInfo.PollingCode {
		if err := transfercoderepo.DeclineConsentByState(rlog, nil, oState); err != nil {
			rlog.Errorf("%s", errorfmt.Full(err))
		}
	}
	return model.Response{
		Status: httpStatus.StatusOKForward,
		Response: map[string]string{
			"url": url,
		},
	}.Send(ctx)
}

// handleConsentAccept handles the acceptance of a consent code
func handleConsentAccept(
	rlog log.Ext1FieldLogger, req *pkg.ConsentApprovalRequest,
	oState *state.State,
) *model.Response {
	for _, c := range req.Capabilities {
		if !api.AllCapabilities.Has(c) {
			return &model.Response{
				Status:   fiber.StatusBadRequest,
				Response: model2.BadRequestError(fmt.Sprintf("unknown capability '%s'", c)),
			}
		}
	}
	for _, c := range req.SubtokenCapabilities {
		if !api.AllCapabilities.Has(c) {
			return &model.Response{
				Status:   fiber.StatusBadRequest,
				Response: model2.BadRequestError(fmt.Sprintf("unknown subtoken_capability '%s'", c)),
			}
		}
	}
	provider, ok := config.Get().ProviderByIssuer[req.Issuer]
	if !ok {
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}
	}
	var authURI string
	var err error
	if err = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err = authcodeinforepo.UpdateTokenInfoByState(
				rlog, tx, oState, req.Restrictions, req.Capabilities, req.SubtokenCapabilities, req.Rotation,
				req.TokenName,
			); err != nil {
				return err
			}
			authURI, err = authcode.GetAuthorizationURL(rlog, tx, provider, oState, req.Restrictions)
			return err
		},
	); err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{
		Status: httpStatus.StatusOKForward,
		Response: map[string]string{
			"authorization_uri": authURI,
		},
	}
}

// HandleConsentPost handles consent confirmation requests
func HandleConsentPost(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	authInfo, oState, err := getAuthInfoFromConsentCodeStr(rlog, ctx.Params("consent_code"))
	if err != nil {
		// Don't log error here, it was already logged
		return err
	}
	if len(ctx.Body()) == 0 {
		return handleConsentDecline(ctx, authInfo, oState)
	}
	req := pkg.ConsentApprovalRequest{}
	if err = json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	return handleConsentAccept(rlog, &req, oState).Send(ctx)
}
