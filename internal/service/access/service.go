// Package access provides a service layer for access token operations.
package access

import (
	"strings"
	"time"

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
	notifier "github.com/oidc-mytoken/server/internal/notifier/client"
	"github.com/oidc-mytoken/server/internal/oidc/oidcreqres"
	"github.com/oidc-mytoken/server/internal/oidc/refresh"
	"github.com/oidc-mytoken/server/internal/utils"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

// Service is the access token service singleton.
var Service = &service{}

type service struct{}

// CreateAccessToken creates an access token from a mytoken.
func (s *service) CreateAccessToken(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData, req request.AccessTokenRequest,
) *model.Response {
	req.GrantType = model.GrantTypeMytoken
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestriction(
				rlog, tx, mt, clientMetaData,
				utils.SplitIgnoreEmpty(req.Scope, " "),
				utils.SplitIgnoreEmpty(req.Audience, " "),
				api.CapabilityAT,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			provider, errRes := auth.RequireMatchingIssuer(rlog, mt.OIDCIssuer, &req.Issuer)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			var rollback bool
			res, rollback = s.createAccessTokenLogic(rlog, tx, mt, req, *clientMetaData, provider, usedRestriction)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

func (s *service) createAccessTokenLogic(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mt *mytoken.Mytoken, req request.AccessTokenRequest,
	networkData api.ClientMetaData,
	provider model.Provider, usedRestriction *restrictions.Restriction,
) (*model.Response, bool) {
	var tokenUpdate *response.MytokenResponse
	var oidcRes *oidcreqres.OIDCTokenResponse
	var rsp *request.AccessTokenResponse

	scopes, auds := s.parseScopesAndAudienceToUse(
		req.Scope, strings.Split(req.Audience, " "), usedRestriction, provider.Scopes(),
	)

	eventComment := "Used grant_type mytoken"

	// Try to serve the request from a cached access token
	if atCacheConf := provider.AccessTokenCache(); atCacheConf != nil && atCacheConf.Enabled() {
		stJWT, err := mt.ToJWT()
		if err != nil {
			rlog.Errorf("%s", errorfmt.Full(err))
			return model.ErrorToInternalServerErrorResponse(err), true
		}
		reqScopes := utils.SplitIgnoreEmpty(scopes, " ")
		cached, err := accesstokenrepo.GetCachedAT(
			rlog, tx, mt.ID, stJWT, reqScopes, auds,
		)
		if err != nil {
			rlog.Errorf("%s", errorfmt.Full(err))
			return model.ErrorToInternalServerErrorResponse(err), true
		}
		if cached != nil && atCacheConf.ShouldReuse(time.Now(), cached.Created, cached.ExpiresAt) {
			rlog.Debug("Returning cached access token")
			eventComment += "; returned a cached access token"
			cachedAudiences, _ := jwtutils.GetAudiencesFromJWT(rlog, cached.Token)
			rsp = &request.AccessTokenResponse{
				AccessTokenResponse: api.AccessTokenResponse{
					AccessToken: cached.Token,
					TokenType:   cached.TokenType,
					ExpiresIn:   max(int64(time.Until(cached.ExpiresAt).Seconds()), 0),
					Scope:       strings.Join(reqScopes, " "),
					Audiences:   cachedAudiences,
				},
			}
		}
	}

	if rsp == nil { // no cached access token could be used, so we obtain a fresh one
		rt, rtFound, dbErr := cryptstore.GetRefreshToken(rlog, tx, mt.ID, req.Mytoken.JWT)
		if dbErr != nil {
			rlog.Errorf("%s", errorfmt.Full(dbErr))
			return model.ErrorToInternalServerErrorResponse(dbErr), true
		}
		if !rtFound {
			_ = notifier.SendNotificationsForSubClass(
				rlog, tx, mt.ID, api.NotificationClassRTFailure, &networkData,
				model.KeyValues{
					{
						Key:   "Reason",
						Value: "No refresh token attached",
					},
				}, nil,
			)
			return &model.Response{
				Status:   fiber.StatusUnauthorized,
				Response: model.InvalidTokenError("No refresh token attached"),
			}, true
		}

		opRes, oidcErrRes, err := refresh.DoFlowAndUpdateDB(
			rlog, tx, provider, mt.ID, req.Mytoken.JWT, rt, scopes, auds,
		)
		if err != nil {
			rlog.Errorf("%s", errorfmt.Full(err))
			return model.ErrorToInternalServerErrorResponse(err), true
		}
		if oidcErrRes != nil {
			_ = notifier.SendNotificationsForSubClass(
				rlog, tx, mt.ID, api.NotificationClassRTFailure, &networkData,
				model.KeyValues{
					{
						Key:   "OP Error",
						Value: oidcErrRes.Error,
					},
					{
						Key:   "OP Error Description",
						Value: oidcErrRes.ErrorDescription,
					},
				}, nil,
			)
			return &model.Response{
				Status:   oidcErrRes.Status,
				Response: model.OIDCError(oidcErrRes.Error, oidcErrRes.ErrorDescription),
			}, true
		}
		oidcRes = opRes

		retScopes := oidcRes.Scopes
		if retScopes == "" {
			retScopes = scopes
		}
		retAudiences, _ := jwtutils.GetAudiencesFromJWT(rlog, oidcRes.AccessToken)
		at := accesstokenrepo.AccessToken{
			Token:     oidcRes.AccessToken,
			IP:        networkData.IP,
			Comment:   req.Comment,
			Mytoken:   mt,
			Scopes:    utils.SplitIgnoreEmpty(retScopes, " "),
			Audiences: retAudiences,
			ExpiresAt: oidcRes.AccessTokenExpiresAt(rlog),
			TokenType: oidcRes.TokenType,
		}

		if err = at.Store(rlog, tx); err != nil {
			rlog.Errorf("%s", errorfmt.Full(err))
			return model.ErrorToInternalServerErrorResponse(err), true
		}

		rsp = &request.AccessTokenResponse{
			AccessTokenResponse: api.AccessTokenResponse{
				AccessToken: oidcRes.AccessToken,
				TokenType:   oidcRes.TokenType,
				ExpiresIn:   oidcRes.ExpiresIn,
				Scope:       retScopes,
				Audiences:   retAudiences,
			},
		}
	}

	if err := eventService.LogEvent(
		rlog, tx, pkg.MTEvent{
			Event:          api.EventATCreated,
			Comment:        eventComment,
			MTID:           mt.ID,
			ClientMetaData: networkData,
		},
	); err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err), true
	}
	if usedRestriction != nil {
		if err := usedRestriction.UsedAT(rlog, tx, mt.ID); err != nil {
			rlog.Errorf("%s", errorfmt.Full(err))
			return model.ErrorToInternalServerErrorResponse(err), true
		}
	}
	var err error
	tokenUpdate, err = rotation.RotateMytokenAfterATForResponse(
		rlog, tx, req.Mytoken.JWT, mt, networkData, req.Mytoken.OriginalTokenType,
	)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err), true
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
	}, false
}

func (s *service) parseScopesAndAudienceToUse(
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
	auds = utils.RemoveEmpty(auds)
	return scopes, auds
}
