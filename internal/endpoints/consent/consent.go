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
	"github.com/oidc-mytoken/server/internal/endpoints/webentities"
	"github.com/oidc-mytoken/server/internal/model/profiled"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	provider2 "github.com/oidc-mytoken/server/internal/oidc/provider"
	"github.com/oidc-mytoken/server/internal/server/httpstatus"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/internal/endpoints/consent/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/authcode"
	"github.com/oidc-mytoken/server/internal/utils"
)

func getAuthInfoFromConsentCodeStr(rlog log.Ext1FieldLogger, code string) (
	*authcodeinforepo.AuthFlowInfoOut, *state.State, error,
) {
	consentCode := state.ConsentCodeFromStr(code)
	oState := state.NewState(consentCode.GetState())
	authInfo, err := authcodeinforepo.GetAuthFlowInfoByState(rlog, nil, oState)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fiber.ErrNotFound
		} else {
			rlog.Errorf("%s", errorfmt.Full(err))
		}
	}
	return authInfo, oState, err
}

// HandleCreateConsent returns consent data as JSON for the posted parameters
func HandleCreateConsent(ctx *fiber.Ctx) error {
	req := pkg2.NewMytokenRequest()
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	if req.Issuer == "" {
		return model.BadRequestErrorResponse("required parameter 'oidc_issuer' missing").Send(ctx)
	}
	rlog := logger.GetRequestLogger(ctx)
	mt, _ := auth.RequireValidMytoken(rlog, nil, &req.Mytoken, ctx)
	r, _ := restrictions.Tighten(rlog, mt.Restrictions, req.Restrictions.Restrictions)
	c := api.TightenCapabilities(mt.Capabilities, req.Capabilities.Capabilities)
	info := &pkg2.OIDCFlowRequest{
		GeneralMytokenRequest: profiled.GeneralMytokenRequest{
			GeneralMytokenRequest: req.GeneralMytokenRequest.GeneralMytokenRequest,
			Capabilities:          profiled.Capabilities{Capabilities: c},
			Restrictions:          profiled.Restrictions{Restrictions: r},
		},
	}
	info.Rotation = req.Rotation
	return handleConsentJSON(ctx, info)
}

// ConsentInfoResponse is the JSON response for consent data
type ConsentInfoResponse struct {
	Issuer          string                       `json:"oidc_issuer"`
	ApplicationName string                       `json:"application_name,omitempty"`
	TokenName       string                       `json:"token_name,omitempty"`
	Capabilities    api.Capabilities             `json:"capabilities"`
	Restrictions    restrictions.Restrictions    `json:"restrictions,omitempty"`
	Rotation        *api.Rotation                `json:"rotation,omitempty"`
	Tags            []api.CreateMytokenTag       `json:"tags,omitempty"`
	SupportedScopes []string                     `json:"supported_scopes,omitempty"`
	AllCapabilities []*webentities.WebCapability `json:"all_capabilities"`
}

// SPAHandler is set by the server package when SPA is enabled
// This allows the consent handler to delegate to the SPA for HTML requests
var SPAHandler fiber.Handler

// HandleConsent displays a consent page or returns JSON for SPA
func HandleConsent(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	authInfo, _, err := getAuthInfoFromConsentCodeStr(rlog, ctx.Params("consent_code"))
	if err != nil {
		// Don't log error here, it was already logged
		return err
	}

	// Check if client accepts JSON (SPA API request)
	accept := ctx.Get("Accept")
	if strings.Contains(accept, "application/json") {
		return handleConsentJSON(ctx, &(authInfo.AuthCodeFlowRequest.OIDCFlowRequest))
	}

	// Serve the SPA for HTML requests (let client-side routing handle it)
	if SPAHandler != nil {
		return SPAHandler(ctx)
	}

	// If SPA is not available, return JSON
	return handleConsentJSON(ctx, &(authInfo.AuthCodeFlowRequest.OIDCFlowRequest))
}

// HandleConsentAPI returns consent data as JSON for SPA API requests
func HandleConsentAPI(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	authInfo, _, err := getAuthInfoFromConsentCodeStr(rlog, ctx.Params("consent_code"))
	if err != nil {
		return err
	}
	return handleConsentJSON(ctx, &(authInfo.AuthCodeFlowRequest.OIDCFlowRequest))
}

// handleConsentJSON returns consent data as JSON for SPA
func handleConsentJSON(ctx *fiber.Ctx, info *pkg2.OIDCFlowRequest) error {
	var scopes []string
	if p := provider2.GetProvider(info.Issuer); p != nil {
		scopes = p.Scopes()
	}

	var rotation *api.Rotation
	if info.Rotation != nil {
		rotation = &info.Rotation.Rotation
	}

	response := ConsentInfoResponse{
		Issuer:          info.Issuer,
		ApplicationName: info.ApplicationName,
		TokenName:       info.Name,
		Capabilities:    info.Capabilities.Capabilities,
		Restrictions:    info.Restrictions.Restrictions,
		Rotation:        rotation,
		Tags:            info.Tags,
		SupportedScopes: scopes,
		AllCapabilities: webentities.AllWebCapabilities(),
	}

	return ctx.JSON(response)
}

func handleConsentDecline(
	ctx *fiber.Ctx, authInfo *authcodeinforepo.AuthFlowInfoOut,
	oState *state.State,
) *model.Response {
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
		return res
	}
	if authInfo.PollingCode {
		if err := transfercoderepo.DeclineConsentByState(rlog, nil, oState); err != nil {
			rlog.Errorf("%s", errorfmt.Full(err))
		}
	}
	return &model.Response{
		Status: httpstatus.StatusOKForward,
		Response: map[string]string{
			"url": url,
		},
	}
}

// handleConsentAccept handles the acceptance of a consent code
func handleConsentAccept(
	rlog log.Ext1FieldLogger, req *pkg.ConsentApprovalRequest,
	oState *state.State,
) *model.Response {
	for _, c := range req.Capabilities {
		if !api.AllCapabilities.Has(c) {
			return model.BadRequestErrorResponse(fmt.Sprintf("unknown capability '%s'", c))
		}
	}
	p := provider2.GetProvider(req.Issuer)
	if p == nil {
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
				rlog, tx, oState, req.Restrictions, req.Capabilities, req.Rotation, req.TokenName, req.Tags,
			); err != nil {
				return err
			}
			authURI, err = authcode.GetAuthorizationURL(rlog, tx, p, oState, req.Restrictions)
			return err
		},
	); err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{
		Status: httpstatus.StatusOKForward,
		Response: map[string]string{
			"authorization_uri": authURI,
		},
	}
}

// HandleConsentPost handles consent confirmation requests
func HandleConsentPost(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	authInfo, oState, err := getAuthInfoFromConsentCodeStr(rlog, ctx.Params("consent_code"))
	if err != nil {
		// Don't log error here, it was already logged
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if len(ctx.Body()) == 0 {
		return handleConsentDecline(ctx, authInfo, oState)
	}
	req := pkg.ConsentApprovalRequest{}
	if err = json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	return handleConsentAccept(rlog, &req, oState)
}
