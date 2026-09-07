// Package tag provides a service layer for tag CRUD operations.
// It handles auth checks, DB operations, and post-processing (event logging,
// restriction tracking, token rotation) in a single call, removing duplication
// between HTTP API and SSH handlers.
package tag

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/tagrepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/mytokenutils"
)

// tagListResponse wraps the tag list response to support token updates.
type tagListResponse struct {
	Tags        []api.TagInfo             `json:"tags"`
	TokenUpdate *response.MytokenResponse `json:"token_update,omitempty"`
}

func (r *tagListResponse) SetTokenUpdate(tu *response.MytokenResponse) {
	r.TokenUpdate = tu
}

// Service is the tag service singleton.
var Service = &service{}

type service struct{}

// List returns all tags for the user identified by the given mytoken.
func (s *service) List(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData,
) *model.Response {
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, clientMetaData, api.CapabilityTagsRead,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			tags, err := tagrepo.ListTags(rlog, tx, mt.ID)
			if err != nil {
				return err
			}
			res = &model.Response{
				Status:   http.StatusOK,
				Response: &tagListResponse{Tags: tags},
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, res, mt, *clientMetaData,
				api.EventTagsListed, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
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

// Create creates a new tag for the user.
func (s *service) Create(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, tag string, color string,
) *model.Response {
	if tag == "" {
		return model.BadRequestErrorResponse("tag must not be empty")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, clientMetaData, api.CapabilityTags,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if err := tagrepo.CreateTag(rlog, tx, tag, color, mt.ID); err != nil {
				return err
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				api.EventTagCreated, tag, usedRestriction, umt.JWT, umt.OriginalTokenType,
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

// Update updates an existing tag.
func (s *service) Update(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, tag string, req api.TagInfo,
) *model.Response {
	if tag == "" {
		return model.BadRequestErrorResponse("tag must not be empty")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, clientMetaData, api.CapabilityTags,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if err := tagrepo.UpdateTag(rlog, tx, tag, req, mt.ID); err != nil {
				return err
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				api.EventTagUpdated, tag, usedRestriction, umt.JWT, umt.OriginalTokenType,
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

// Delete deletes a tag.
func (s *service) Delete(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, tag string,
) *model.Response {
	if tag == "" {
		return model.BadRequestErrorResponse("tag must not be empty")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, clientMetaData, api.CapabilityTags,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if err := tagrepo.DeleteTag(rlog, tx, tag, mt.ID); err != nil {
				return err
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				api.EventTagDeleted, tag, usedRestriction, umt.JWT, umt.OriginalTokenType,
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
