package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/internal/utils/mytokenutils"
)

type addTagSSHRequest struct {
	Tag             string `json:"tag"`
	MomID           string `json:"mom_id"`
	IncludeChildren bool   `json:"include_children"`
}

type removeTagSSHRequest struct {
	Tag   string `json:"tag"`
	MomID string `json:"mom_id"`
}

func handleSSHAddTag(reqData []byte, s ssh.Session) error {
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle add-tag from ssh")

	var req addTagSSHRequest
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.Tag == "" {
		return errors.New("tag is required")
	}

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	momID := mt.ID.MomID()
	if req.MomID != "" {
		momID = mtid.MOMID{MTID: mtid.FromHash(req.MomID)}
	}
	id, momMode, errRes := auth.ValidateCapabilityWithMomMode(
		rlog, api.CapabilityTokeninfoTags,
		api.CapabilityTagAnyToken, mt, momID, clientMetaData,
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}

	umt := mt.ToUniversalMytoken()
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err := mytokenrepo.AddTag(rlog, tx, id.MomID(), api.Tag(req.Tag), req.IncludeChildren); err != nil {
				return err
			}
			var rollback bool
			event := api.EventTagAddedToken
			if momMode {
				event = api.EventTagAddedTokenOther
			}
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				event, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	)
	if err != nil && res == nil {
		res = model.ErrorToInternalServerErrorResponse(err)
	}
	if res != nil {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHRemoveTag(reqData []byte, s ssh.Session) error {
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle remove-tag from ssh")

	var req removeTagSSHRequest
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.Tag == "" {
		return errors.New("tag is required")
	}

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	momID := mt.ID.MomID()
	if req.MomID != "" {
		momID = mtid.MOMID{MTID: mtid.FromHash(req.MomID)}
	}
	id, momMode, errRes := auth.ValidateCapabilityWithMomMode(
		rlog, api.CapabilityTokeninfoTags,
		api.CapabilityTagAnyToken, mt, momID, clientMetaData,
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	usedRestriction, errRes := auth.RequireUsableRestrictionOther(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}

	umt := mt.ToUniversalMytoken()
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err := mytokenrepo.RemoveTag(rlog, tx, id.MomID(), api.Tag(req.Tag)); err != nil {
				return err
			}
			var rollback bool
			event := api.EventTagRemovedToken
			if momMode {
				event = api.EventTagRemovedTokenOther
			}
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				event, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	)
	if err != nil && res == nil {
		res = model.ErrorToInternalServerErrorResponse(err)
	}
	if res != nil {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}
