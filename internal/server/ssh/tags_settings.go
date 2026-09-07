package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/service/tag"
)

func handleSSHTagsList(s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle tags-list from ssh")
	res := tag.Service.List(c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData)
	if res == nil {
		return writeError(s, errors.New("internal server error"))
	}
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeJSON(s, res.Response)
}

func handleSSHTagCreate(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle tag-create from ssh")

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
	color := ""
	if req.Color != nil {
		color = *req.Color
	}
	res := tag.Service.Create(c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData, req.Tag, color)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHTagUpdate(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle tag-update from ssh")

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

	var tagReq api.TagInfo
	if req.Color != nil {
		tagReq.Color = *req.Color
	}
	if req.Name != nil {
		tagReq.Tag = api.Tag(*req.Name)
	}
	res := tag.Service.Update(c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData, req.Tag, tagReq)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}

func handleSSHTagDelete(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle tag-delete from ssh")

	var req struct {
		Tag string `json:"tag"`
	}
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.Tag == "" {
		return errors.New("tag is required")
	}
	res := tag.Service.Delete(c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData, req.Tag)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}
