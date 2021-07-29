package access

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/accesstokenrepo"
	dbhelper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/refreshtokenrepo"
	request "github.com/oidc-mytoken/server/internal/endpoints/token/access/pkg"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	serverModel "github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/refresh"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/ctxUtils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/shared/model"
	eventService "github.com/oidc-mytoken/server/shared/mytoken/event"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
	"github.com/oidc-mytoken/server/shared/mytoken/rotation"
	"github.com/oidc-mytoken/server/shared/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/shared/utils"
	"github.com/oidc-mytoken/server/shared/utils/jwtutils"
)

// HandleAccessTokenEndpoint handles request on the access token endpoint
func HandleAccessTokenEndpoint(ctx *fiber.Ctx) error {
	log.Debug("Handle access token request")
	req := request.NewAccessTokenRequest()
	if err := ctx.BodyParser(&req); err != nil {
		return serverModel.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	log.Trace("Parsed access token request")
	if req.Mytoken.JWT == "" {
		req.Mytoken = req.RefreshToken
	}

	if req.GrantType != model.GrantTypeMytoken {
		res := serverModel.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnsupportedGrantType,
		}
		return res.Send(ctx)
	}
	log.Trace("Checked grant type")
	if req.Mytoken.JWT == "" {
		var err error
		req.Mytoken, err = universalmytoken.Parse(ctx.Cookies("mytoken"))
		if err != nil {
			return serverModel.Response{
				Status:   fiber.StatusUnauthorized,
				Response: model.InvalidTokenError(errorfmt.Error(err)),
			}.Send(ctx)
		}
	}

	mt, err := mytoken.ParseJWT(req.Mytoken.JWT)
	if err != nil {
		return (&serverModel.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.InvalidTokenError(errorfmt.Error(err)),
		}).Send(ctx)
	}
	log.Trace("Parsed mytoken")

	revoked, dbErr := dbhelper.CheckTokenRevoked(nil, mt.ID, mt.SeqNo, mt.Rotation)
	if dbErr != nil {
		log.Errorf("%s", errorfmt.Full(dbErr))
		return serverModel.ErrorToInternalServerErrorResponse(dbErr).Send(ctx)
	}
	if revoked {
		return (&serverModel.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.InvalidTokenError(""),
		}).Send(ctx)
	}
	log.Trace("Checked token not revoked")

	if ok := mt.Restrictions.VerifyForAT(nil, ctx.IP(), mt.ID); !ok {
		return (&serverModel.Response{
			Status:   fiber.StatusForbidden,
			Response: api.ErrorUsageRestricted,
		}).Send(ctx)
	}
	log.Trace("Checked mytoken restrictions")
	if ok := mt.VerifyCapabilities(api.CapabilityAT); !ok {
		res := serverModel.Response{
			Status:   fiber.StatusForbidden,
			Response: api.ErrorInsufficientCapabilities,
		}
		return res.Send(ctx)
	}
	log.Trace("Checked mytoken capabilities")
	if req.Issuer == "" {
		req.Issuer = mt.OIDCIssuer
	} else if req.Issuer != mt.OIDCIssuer {
		res := serverModel.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError("token not for specified issuer"),
		}
		return res.Send(ctx)
	}
	log.Trace("Checked issuer")

	return handleAccessTokenRefresh(mt, req, *ctxUtils.ClientMetaData(ctx)).Send(ctx)
}

func handleAccessTokenRefresh(mt *mytoken.Mytoken, req request.AccessTokenRequest, networkData api.ClientMetaData) *serverModel.Response {
	provider, ok := config.Get().ProviderByIssuer[req.Issuer]
	if !ok {
		return &serverModel.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnknownIssuer,
		}
	}

	scopes := strings.Join(provider.Scopes, " ") // default if no restrictions apply
	auds := ""                                   // default if no restrictions apply
	var usedRestriction *restrictions.Restriction
	if len(mt.Restrictions) > 0 {
		possibleRestrictions := mt.Restrictions.GetValidForAT(nil, networkData.IP, mt.ID).
			WithScopes(utils.SplitIgnoreEmpty(req.Scope, " ")).
			WithAudiences(utils.SplitIgnoreEmpty(req.Audience, " "))
		if len(possibleRestrictions) == 0 {
			return &serverModel.Response{
				Status:   fiber.StatusForbidden,
				Response: api.ErrorUsageRestricted,
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
	rt, rtFound, dbErr := refreshtokenrepo.GetRefreshToken(nil, mt.ID, req.Mytoken.JWT)
	if dbErr != nil {
		log.Errorf("%s", errorfmt.Full(dbErr))
		return serverModel.ErrorToInternalServerErrorResponse(dbErr)
	}
	if !rtFound {
		return &serverModel.Response{
			Status:   fiber.StatusUnauthorized,
			Response: model.InvalidTokenError("No refresh token attached"),
		}
	}

	oidcRes, oidcErrRes, err := refresh.RefreshFlowAndUpdateDB(provider, mt.ID, req.Mytoken.JWT, rt, scopes, auds)
	if err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		return serverModel.ErrorToInternalServerErrorResponse(err)
	}
	if oidcErrRes != nil {
		return &serverModel.Response{
			Status:   oidcErrRes.Status,
			Response: model.OIDCError(oidcErrRes.Error, oidcErrRes.ErrorDescription),
		}
	}
	retScopes := oidcRes.Scopes
	if retScopes == "" {
		retScopes = scopes
	}
	retAudiences, _ := jwtutils.GetAudiencesFromJWT(oidcRes.AccessToken)
	at := accesstokenrepo.AccessToken{
		Token:     oidcRes.AccessToken,
		IP:        networkData.IP,
		Comment:   req.Comment,
		Mytoken:   mt,
		Scopes:    utils.SplitIgnoreEmpty(retScopes, " "),
		Audiences: retAudiences,
	}
	var tokenUpdate *response.MytokenResponse
	if err = db.Transact(func(tx *sqlx.Tx) error {
		if err = at.Store(tx); err != nil {
			return err
		}
		if err = eventService.LogEvent(tx, eventService.MTEvent{
			Event: event.FromNumber(event.MTEventATCreated, "Used grant_type mytoken"),
			MTID:  mt.ID,
		}, networkData); err != nil {
			return err
		}
		if usedRestriction != nil {
			if err = usedRestriction.UsedAT(tx, mt.ID); err != nil {
				return err
			}
		}
		tokenUpdate, err = rotation.RotateMytokenAfterATForResponse(
			tx, req.Mytoken.JWT, mt, networkData, req.Mytoken.OriginalTokenType)
		return err
	}); err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		return serverModel.ErrorToInternalServerErrorResponse(err)
	}

	rsp := request.AccessTokenResponse{
		AccessTokenResponse: api.AccessTokenResponse{
			AccessToken: oidcRes.AccessToken,
			TokenType:   oidcRes.TokenType,
			ExpiresIn:   oidcRes.ExpiresIn,
			Scope:       retScopes,
			Audiences:   retAudiences,
		},
	}
	var cake []*fiber.Cookie
	if tokenUpdate != nil {
		rsp.TokenUpdate = tokenUpdate
		cookie := cookies.MytokenCookie(tokenUpdate.Mytoken)
		cake = []*fiber.Cookie{&cookie}
	}
	return &serverModel.Response{
		Status:   fiber.StatusOK,
		Response: rsp,
		Cookies:  cake,
	}
}
