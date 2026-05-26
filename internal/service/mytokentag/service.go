// Package mytokentag provides a service layer for adding/removing tags to/from mytokens.
package mytokentag

import (
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/mytokenutils"
)

// Service is the mytoken tag service singleton.
var Service = &service{}

type service struct{}

// AddTag adds a tag to a mytoken.
func (s *service) AddTag(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, tag string, momID mtid.MOMID,
	includeChildren bool,
) *model.Response {
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			id, momMode, errRes := auth.ValidateCapabilityWithMomMode(
				rlog, tx, api.CapabilityTokeninfoTags, api.CapabilityTagAnyToken,
				mt, momID, clientMetaData,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, tx, mt, clientMetaData)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if err := mytokenrepo.AddTag(rlog, tx, id.MomID(), api.Tag(tag), includeChildren); err != nil {
				return err
			}
			event := api.EventTagAddedToken
			if momMode {
				event = api.EventTagAddedTokenOther
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				event, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
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

// RemoveTag removes a tag from a mytoken.
func (s *service) RemoveTag(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, tag string, momID mtid.MOMID,
) *model.Response {
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			id, momMode, errRes := auth.ValidateCapabilityWithMomMode(
				rlog, tx, api.CapabilityTokeninfoTags, api.CapabilityTagAnyToken,
				mt, momID, clientMetaData,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, tx, mt, clientMetaData)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if err := mytokenrepo.RemoveTag(rlog, tx, id.MomID(), api.Tag(tag)); err != nil {
				return err
			}
			event := api.EventTagRemovedToken
			if momMode {
				event = api.EventTagRemovedTokenOther
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				event, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
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
