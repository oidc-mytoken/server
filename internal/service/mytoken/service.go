// Package mytoken provides a service layer for creating mytokens from existing mytokens.
package mytoken

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/unixtime"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/encryptionkeyrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/refreshtokenrepo"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	"github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/rotation"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

// Service is the mytoken service singleton.
var Service = &service{}

type service struct{}

// CreateFromMytoken creates a new mytoken from an existing mytoken.
func (s *service) CreateFromMytoken(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, req *response.MytokenFromMytokenRequest,
) *model.Response {
	req.GrantType = model.GrantTypeMytoken
	req.Restrictions.ReplaceThisIP(clientMetaData.IP)
	req.Restrictions.ClearUnsupportedKeys()
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, clientMetaData, api.CapabilityCreateMT,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if _, errRes = auth.RequireMatchingIssuer(
				rlog, mt.OIDCIssuer, &req.GeneralMytokenRequest.Issuer,
			); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			var rollback bool
			res, rollback = s.createFromMytokenLogic(rlog, tx, mt, req, clientMetaData, usedRestriction)
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

func (s *service) createFromMytokenLogic(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, parent *mytoken.Mytoken, req *response.MytokenFromMytokenRequest,
	networkData *api.ClientMetaData,
	usedRestriction *restrictions.Restriction,
) (*model.Response, bool) {
	ste, errorResponse := s.createMytokenEntry(rlog, parent, req, *networkData)
	if errorResponse != nil {
		return errorResponse, true
	}
	tokenUpdate, err := s.processSubtokenCreation(rlog, tx, parent, ste, req, networkData, usedRestriction)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err), true
	}

	return s.buildMytokenResponse(rlog, ste, req, networkData, tokenUpdate), false
}

func (s *service) processSubtokenCreation(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, parent *mytoken.Mytoken, ste *mytokenrepo.MytokenEntry,
	req *response.MytokenFromMytokenRequest, networkData *api.ClientMetaData,
	usedRestriction *restrictions.Restriction,
) (*response.MytokenResponse, error) {
	if err := s.markRestrictionUsed(rlog, tx, usedRestriction, parent.ID); err != nil {
		return nil, err
	}

	tokenUpdate, err := rotation.RotateMytokenAfterOtherForResponse(
		rlog, tx, req.Mytoken.JWT, parent, *networkData, req.Mytoken.OriginalTokenType,
	)
	if err != nil {
		return nil, err
	}

	if err = s.storeSubtokenWithInheritance(rlog, tx, parent.ID, ste, req); err != nil {
		return nil, err
	}

	if err = s.logSubtokenEvents(rlog, tx, parent.ID, ste.ID, req.GeneralMytokenRequest.Name, networkData); err != nil {
		return nil, err
	}

	return tokenUpdate, nil
}

func (s *service) markRestrictionUsed(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, r *restrictions.Restriction, mtID mtid.MTID,
) error {
	if r == nil {
		return nil
	}
	return r.UsedOther(rlog, tx, mtID)
}

func (s *service) storeSubtokenWithInheritance(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, parentID mtid.MTID, ste *mytokenrepo.MytokenEntry,
	req *response.MytokenFromMytokenRequest,
) error {
	if err := ste.Store(rlog, tx, "Used grant_type mytoken"); err != nil {
		return err
	}
	if err := notificationsrepo.ExpandNotificationsToChildrenIfApplicable(rlog, tx, parentID, ste.ID); err != nil {
		return err
	}
	if err := mytokenrepo.ExpandTagsToChildrenIfApplicable(rlog, tx, parentID, ste.ID); err != nil {
		return err
	}
	for _, sub := range req.SubscribeNotificationRequests {
		if err := notificationsrepo.MytokenSubscribeOrCreateNotificationWithClasses(rlog, tx, sub, ste.ID); err != nil {
			return err
		}
	}
	return notificationsrepo.ScheduleExpirationNotificationsIfNeeded(
		rlog, tx, ste.ID, ste.Token.ExpiresAt, ste.Token.IssuedAt,
	)
}

func (s *service) logSubtokenEvents(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, parentID, childID mtid.MTID, tokenName string,
	networkData *api.ClientMetaData,
) error {
	return eventService.LogEvents(
		rlog, tx, []pkg.MTEvent{
			{
				Event:          api.EventInheritedRT,
				Comment:        "Got RT from parent",
				MTID:           childID,
				ClientMetaData: *networkData,
			},
			{
				Event:          api.EventSubtokenCreated,
				Comment:        strings.TrimSpace(fmt.Sprintf("Created MT %s", tokenName)),
				MTID:           parentID,
				ClientMetaData: *networkData,
			},
		},
	)
}

func (s *service) buildMytokenResponse(
	rlog log.Ext1FieldLogger, ste *mytokenrepo.MytokenEntry, req *response.MytokenFromMytokenRequest,
	networkData *api.ClientMetaData, tokenUpdate *response.MytokenResponse,
) *model.Response {
	res, err := ste.Token.ToTokenResponse(
		rlog, req.ResponseType, req.GeneralMytokenRequest.MaxTokenLen, *networkData, "",
	)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	var cake []*fiber.Cookie
	if tokenUpdate != nil {
		res.TokenUpdate = tokenUpdate
		cake = []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)}
	}
	return &model.Response{
		Status:   fiber.StatusOK,
		Response: res,
		Cookies:  cake,
	}
}

func (s *service) createMytokenEntry(
	rlog log.Ext1FieldLogger, parent *mytoken.Mytoken, req *response.MytokenFromMytokenRequest,
	networkData api.ClientMetaData,
) (*mytokenrepo.MytokenEntry, *model.Response) {
	rtID, dbErr := refreshtokenrepo.GetRTID(rlog, nil, parent.ID)
	rtFound, err := db.ParseError(dbErr)
	if err != nil {
		rlog.WithError(dbErr).Error()
		return nil, model.ErrorToInternalServerErrorResponse(dbErr)
	}
	if !rtFound {
		return nil, &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.InvalidTokenError(""),
		}
	}
	if changed := req.Restrictions.EnforceMaxLifetime(parent.OIDCIssuer); changed && req.FailOnRestrictionsNotTighter {
		return nil, model.BadRequestErrorResponse("requested restrictions do not respect maximum mytoken lifetime")
	}
	req.Restrictions.ResolveDefaultAnchors(unixtime.Now())
	r, ok := restrictions.Tighten(rlog, parent.Restrictions, req.Restrictions.Restrictions)
	if !ok && req.FailOnRestrictionsNotTighter {
		return nil, model.BadRequestErrorResponse("requested restrictions are not subset of original restrictions")
	}
	c := api.TightenCapabilities(parent.Capabilities, req.Capabilities.Capabilities)
	if len(c) == 0 {
		return nil, model.BadRequestErrorResponse("mytoken to be issued cannot have any of the requested capabilities")
	}
	var rot *api.Rotation
	if req.Rotation != nil {
		rot = &req.Rotation.Rotation
	}
	mt, err := mytoken.NewMytoken(
		parent.OIDCSubject, parent.OIDCIssuer, req.GeneralMytokenRequest.Name, r, c, rot,
		parent.AuthTime,
	)
	if err != nil {
		return nil, model.ErrorToInternalServerErrorResponse(err)
	}
	mte := mytokenrepo.NewMytokenEntry(mt, req.GeneralMytokenRequest.Name, networkData)
	mte.Tags = req.GeneralMytokenRequest.Tags
	encryptionKey, _, err := encryptionkeyrepo.GetEncryptionKey(rlog, nil, parent.ID, req.Mytoken.JWT)
	if err != nil {
		rlog.WithError(err).Error()
		return mte, model.ErrorToInternalServerErrorResponse(err)
	}
	if err = mte.SetRefreshToken(rtID, encryptionKey); err != nil {
		rlog.WithError(err).Error()
		return mte, model.ErrorToInternalServerErrorResponse(err)
	}
	mte.ParentID = parent.ID
	return mte, nil
}
