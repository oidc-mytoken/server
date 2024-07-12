package access

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils/jwtutils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/accesstokenrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/cryptstore"
	request "github.com/oidc-mytoken/server/internal/endpoints/token/access/pkg"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	"github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/rotation"
	"github.com/oidc-mytoken/server/internal/oidc/oidcreqres"
	"github.com/oidc-mytoken/server/internal/oidc/refresh"
	"github.com/oidc-mytoken/server/internal/utils"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleAccessTokenEndpoint handles request on the access token endpoint
func HandleAccessTokenEndpoint(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle access token request")
	req := request.NewAccessTokenRequest()
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	rlog.Trace("Parsed access token request")
	if req.Mytoken.JWT == "" {
		req.Mytoken = req.RefreshToken
	}

	if errRes := auth.RequireGrantType(rlog, model.GrantTypeMytoken, req.GrantType); errRes != nil {
		return errRes
	}
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &req.Mytoken, ctx)
	if errRes != nil {
		return errRes
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestriction(
		rlog, nil, mt, ctxutils.ClientMetaData(ctx),
		utils.SplitIgnoreEmpty(req.Scope, " "),
		utils.SplitIgnoreEmpty(req.Audience, " "),
		api.CapabilityAT,
	)
	if errRes != nil {
		return errRes
	}
	provider, errRes := auth.RequireMatchingIssuer(rlog, mt.OIDCIssuer, &req.Issuer)
	if errRes != nil {
		return errRes
	}

	return HandleAccessTokenRefresh(rlog, mt, req, *ctxutils.ClientMetaData(ctx), provider, usedRestriction)
}

func parseScopesAndAudienceToUse(
	reqScope string, reqAud []string, usedRestriction *restrictions.Restriction,
	providerScopes []string,
) (
	string,
	[]string,
) {
	scopes := strings.Join(providerScopes, " ") // default if no restrictions apply
	auds := []string{}                          // default if no restrictions apply
	if usedRestriction != nil {
		if reqScope != "" {
			scopes = reqScope
		} else if usedRestriction.Scope != "" {
			scopes = usedRestriction.Scope
		}
		if len(reqAud) > 0 {
			auds = reqAud
		} else if len(usedRestriction.Audiences) > 0 {
			auds = usedRestriction.Audiences
		}
	}
	return scopes, auds
}

// HandleAccessTokenRefresh handles an access token request
func HandleAccessTokenRefresh(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, req request.AccessTokenRequest, networkData api.ClientMetaData,
	provider model.Provider, usedRestriction *restrictions.Restriction,
) *model.Response {
	var errRes *model.Response
	var tokenUpdate *response.MytokenResponse
	var oidcRes *oidcreqres.OIDCTokenResponse
	var retScopes string
	var retAudiences []string
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			rt, rtFound, dbErr := cryptstore.GetRefreshToken(rlog, tx, mt.ID, req.Mytoken.JWT)
			if dbErr != nil {
				return dbErr
			}
			if !rtFound {
				errRes = &model.Response{
					Status:   fiber.StatusUnauthorized,
					Response: model.InvalidTokenError("No refresh token attached"),
				}
				return errors.New("rollback")
			}

			scopes, auds := parseScopesAndAudienceToUse(
				req.Scope, strings.Split(req.Audience, " "), usedRestriction, provider.Scopes(),
			)
			opRes, oidcErrRes, err := refresh.DoFlowAndUpdateDB(
				rlog, tx, provider, mt.ID, req.Mytoken.JWT, rt, scopes, auds,
			)
			if err != nil {
				return err
			}
			if oidcErrRes != nil {
				errRes = &model.Response{
					Status:   oidcErrRes.Status,
					Response: model.OIDCError(oidcErrRes.Error, oidcErrRes.ErrorDescription),
				}
				return errors.New("rollback")
			}
			oidcRes = opRes

			retScopes = oidcRes.Scopes
			if retScopes == "" {
				retScopes = scopes
			}
			retAudiences, _ = jwtutils.GetAudiencesFromJWT(rlog, oidcRes.AccessToken)
			at := accesstokenrepo.AccessToken{
				Token:     oidcRes.AccessToken,
				IP:        networkData.IP,
				Comment:   req.Comment,
				Mytoken:   mt,
				Scopes:    utils.SplitIgnoreEmpty(retScopes, " "),
				Audiences: retAudiences,
			}

			if err = at.Store(rlog, tx); err != nil {
				return err
			}
			if err = eventService.LogEvent(
				rlog, tx, pkg.MTEvent{
					Event:          api.EventATCreated,
					Comment:        "Used grant_type mytoken",
					MTID:           mt.ID,
					ClientMetaData: networkData,
				},
			); err != nil {
				return err
			}
			if usedRestriction != nil {
				if err = usedRestriction.UsedAT(rlog, tx, mt.ID); err != nil {
					return err
				}
			}
			tokenUpdate, err = rotation.RotateMytokenAfterATForResponse(
				rlog, tx, req.Mytoken.JWT, mt, networkData, req.Mytoken.OriginalTokenType,
			)
			return err
		},
	); err != nil {
		if errRes != nil {
			return errRes
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
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
		cake = []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)}
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: rsp,
		Cookies:  cake,
	}
}
