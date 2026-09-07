package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/service/mytokentag"
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
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle add-tag from ssh")

	var req addTagSSHRequest
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.Tag == "" {
		return errors.New("tag is required")
	}
	momID := c.mt.ID.MomID()
	if req.MomID != "" {
		momID = mtid.MOMID{MTID: mtid.FromHash(req.MomID)}
	}
	res := mytokentag.Service.AddTag(
		c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData, req.Tag, momID, req.IncludeChildren,
	)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHRemoveTag(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle remove-tag from ssh")

	var req removeTagSSHRequest
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.Tag == "" {
		return errors.New("tag is required")
	}
	momID := c.mt.ID.MomID()
	if req.MomID != "" {
		momID = mtid.MOMID{MTID: mtid.FromHash(req.MomID)}
	}
	res := mytokentag.Service.RemoveTag(
		c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData, req.Tag, momID,
	)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}
