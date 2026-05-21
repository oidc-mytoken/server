package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/tagrepo"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/internal/utils/mytokenutils"
)

type tagActionRequest struct {
	Tag   string  `json:"tag"`
	Color *string `json:"color,omitempty"`
	Name  *string `json:"name,omitempty"`
}

func handleSSHTagsList(s ssh.Session) error {
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle tags-list from ssh")

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt, clientMetaData, api.CapabilityTagsRead,
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}

	umt := mt.ToUniversalMytoken()
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			tags, err := tagrepo.ListTags(rlog, tx, mt.ID)
			if err != nil {
				return err
			}
			res = &model.Response{
				Status: 200,
				Response: &struct {
					Tags []api.TagInfo `json:"tags"`
				}{Tags: tags},
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
	)
	if err != nil && res == nil {
		return err
	}
	if res != nil {
		return writeJSON(s, res.Response)
	}
	return writeJSON(
		s, struct {
			Tags []api.TagInfo `json:"tags"`
		}{Tags: []api.TagInfo{}},
	)
}

func handleSSHTagCreate(reqData []byte, s ssh.Session) error {
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle tag-create from ssh")

	var req struct {
		Tag   string  `json:"tag"`
		Color *string `json:"color"`
	}
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
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt, clientMetaData, api.CapabilityTags,
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}

	color := ""
	if req.Color != nil {
		color = *req.Color
	}
	umt := mt.ToUniversalMytoken()
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err := tagrepo.CreateTag(rlog, tx, req.Tag, color, mt.ID); err != nil {
				return err
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				api.EventTagCreated, req.Tag, usedRestriction, umt.JWT, umt.OriginalTokenType,
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

func handleSSHTagUpdate(reqData []byte, s ssh.Session) error {
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle tag-update from ssh")

	var req struct {
		Tag   string  `json:"tag"`
		Color *string `json:"color,omitempty"`
		Name  *string `json:"name,omitempty"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.Tag == "" {
		return errors.New("tag is required")
	}
	if (req.Color == nil || *req.Color == "") && (req.Name == nil || *req.Name == "") {
		return errors.New("no supported request parameter given")
	}

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt, clientMetaData, api.CapabilityTags,
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}

	var tagReq api.TagInfo
	if req.Color != nil {
		tagReq.Color = *req.Color
	}
	if req.Name != nil {
		tagReq.Tag = api.Tag(*req.Name)
	}
	umt := mt.ToUniversalMytoken()
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err := tagrepo.UpdateTag(rlog, tx, req.Tag, tagReq, mt.ID); err != nil {
				return err
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				api.EventTagUpdated, req.Tag, usedRestriction, umt.JWT, umt.OriginalTokenType,
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

func handleSSHTagDelete(reqData []byte, s ssh.Session) error {
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle tag-delete from ssh")

	var req struct {
		Tag string `json:"tag"`
	}
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
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt, clientMetaData, api.CapabilityTags,
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}

	umt := mt.ToUniversalMytoken()
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if err := tagrepo.DeleteTag(rlog, tx, req.Tag, mt.ID); err != nil {
				return err
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				api.EventTagDeleted, req.Tag, usedRestriction, umt.JWT, umt.OriginalTokenType,
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
