package access

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/accesstokenrepo"
	dbhelper "github.com/oidc-mytoken/server/internal/db/dbrepo/supertokenrepo/supertokenrepohelper"
	request "github.com/oidc-mytoken/server/internal/endpoints/token/access/pkg"
	serverModel "github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/refresh"
	"github.com/oidc-mytoken/server/internal/utils/ctxUtils"
	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/supertoken/capabilities"
	eventService "github.com/oidc-mytoken/server/shared/supertoken/event"
	event "github.com/oidc-mytoken/server/shared/supertoken/event/pkg"
	supertoken "github.com/oidc-mytoken/server/shared/supertoken/pkg"
	"github.com/oidc-mytoken/server/shared/supertoken/restrictions"
	"github.com/oidc-mytoken/server/shared/supertoken/token"
	"github.com/oidc-mytoken/server/shared/utils"
	"github.com/oidc-mytoken/server/shared/utils/jwtutils"
)

// HandleAccessTokenEndpoint handles request on the access token endpoint
func HandleAccessTokenEndpoint(ctx *fiber.Ctx) error {
	log.Debug("Handle access token request")
	req := request.AccessTokenRequest{}
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return serverModel.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	log.Trace("Parsed access token request")

	if req.GrantType != model.GrantTypeSuperToken {
		res := serverModel.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.APIErrorUnsupportedGrantType,
		}
		return res.Send(ctx)
	}
	log.Trace("Checked grant type")
	if len(req.SuperToken) == 0 {
		var err error
		req.SuperToken, err = token.GetLongSuperToken(ctx.Cookies("mytoken-supertoken"))
		if err != nil {
			return serverModel.Response{
				Status:   fiber.StatusUnauthorized,
				Response: model.InvalidTokenError(err.Error()),
			}.Send(ctx)
		}
	}

	st, err := supertoken.ParseJWT(string(req.SuperToken))
	if err != nil {
		return (&serverModel.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.InvalidTokenError(err.Error()),
		}).Send(ctx)
	}
	log.Trace("Parsed super token")

	revoked, dbErr := dbhelper.CheckTokenRevoked(st.ID)
	if dbErr != nil {
		return serverModel.ErrorToInternalServerErrorResponse(dbErr).Send(ctx)
	}
	if revoked {
		return (&serverModel.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.InvalidTokenError(""),
		}).Send(ctx)
	}
	log.Trace("Checked token not revoked")

	if ok := st.Restrictions.VerifyForAT(nil, ctx.IP(), st.ID); !ok {
		return (&serverModel.Response{
			Status:   fiber.StatusForbidden,
			Response: model.APIErrorUsageRestricted,
		}).Send(ctx)
	}
	log.Trace("Checked super token restrictions")
	if ok := st.VerifyCapabilities(capabilities.CapabilityAT); !ok {
		res := serverModel.Response{
			Status:   fiber.StatusForbidden,
			Response: model.APIErrorInsufficientCapabilities,
		}
		return res.Send(ctx)
	}
	log.Trace("Checked super token capabilities")
	if req.Issuer == "" {
		req.Issuer = st.OIDCIssuer
	} else if req.Issuer != st.OIDCIssuer {
		res := serverModel.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError("token not for specified issuer"),
		}
		return res.Send(ctx)
	}
	log.Trace("Checked issuer")

	return handleAccessTokenRefresh(st, req, *ctxUtils.ClientMetaData(ctx)).Send(ctx)
}

func handleAccessTokenRefresh(st *supertoken.SuperToken, req request.AccessTokenRequest, networkData serverModel.ClientMetaData) *serverModel.Response {
	provider, ok := config.Get().ProviderByIssuer[req.Issuer]
	if !ok {
		return &serverModel.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.APIErrorUnknownIssuer,
		}
	}

	scopes := strings.Join(provider.Scopes, " ") // default if no restrictions apply
	auds := ""                                   // default if no restrictions apply
	var usedRestriction *restrictions.Restriction
	if len(st.Restrictions) > 0 {
		possibleRestrictions := st.Restrictions.GetValidForAT(nil, networkData.IP, st.ID).WithScopes(utils.SplitIgnoreEmpty(req.Scope, " ")).WithAudiences(utils.SplitIgnoreEmpty(req.Audience, " "))
		if len(possibleRestrictions) == 0 {
			return &serverModel.Response{
				Status:   fiber.StatusForbidden,
				Response: model.APIErrorUsageRestricted,
			}
		}
		usedRestriction = &possibleRestrictions[0]
		if req.Scope != "" {
			scopes = req.Scope
		} else if usedRestriction.Scope != "" {
			scopes = usedRestriction.Scope
		}
		if req.Audience != "" {
			auds = req.Audience
		} else if len(usedRestriction.Audiences) > 0 {
			auds = strings.Join(usedRestriction.Audiences, " ")
		}
	}
	rt, rtFound, dbErr := dbhelper.GetRefreshToken(st.ID, string(req.SuperToken))
	if dbErr != nil {
		return serverModel.ErrorToInternalServerErrorResponse(dbErr)
	}
	if !rtFound {
		return &serverModel.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.InvalidTokenError("No refresh token attached"),
		}
	}

	oidcRes, oidcErrRes, err := refresh.RefreshFlowAndUpdateDB(provider, st.ID, string(req.SuperToken), rt, scopes, auds)
	if err != nil {
		return serverModel.ErrorToInternalServerErrorResponse(err)
	}
	if oidcErrRes != nil {
		return &serverModel.Response{
			Status:   oidcErrRes.Status,
			Response: model.OIDCError(oidcErrRes.Error, oidcErrRes.ErrorDescription),
		}
	}
	retScopes := scopes
	if oidcRes.Scopes != "" {
		retScopes = oidcRes.Scopes
	}
	retAudiences, _ := jwtutils.GetAudiencesFromJWT(oidcRes.AccessToken)
	at := accesstokenrepo.AccessToken{
		Token:      oidcRes.AccessToken,
		IP:         networkData.IP,
		Comment:    req.Comment,
		SuperToken: st,
		Scopes:     utils.SplitIgnoreEmpty(retScopes, " "),
		Audiences:  retAudiences,
	}
	if err = db.Transact(func(tx *sqlx.Tx) error {
		if err = at.Store(tx); err != nil {
			return err
		}
		if err = eventService.LogEvent(tx, event.FromNumber(event.STEventATCreated, "Used grant_type super_token"), st.ID, networkData); err != nil {
			return err
		}
		if usedRestriction != nil {
			if err = usedRestriction.UsedAT(tx, st.ID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return serverModel.ErrorToInternalServerErrorResponse(err)
	}
	return &serverModel.Response{
		Status: fiber.StatusOK,
		Response: request.AccessTokenResponse{
			AccessToken: oidcRes.AccessToken,
			TokenType:   oidcRes.TokenType,
			ExpiresIn:   oidcRes.ExpiresIn,
			Scope:       retScopes,
			Audiences:   retAudiences,
		},
	}
}
