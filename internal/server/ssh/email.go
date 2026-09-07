package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/service/email"
)

func handleSSHEmailGet(s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle email-get from ssh")
	res := email.Service.Get(c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData)
	if res == nil {
		return writeError(s, errors.New("internal server error"))
	}
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeJSON(s, res.Response)
}

func handleSSHEmailSet(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle email-set from ssh")

	var req api.UpdateMailSettingsRequest
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.PreferHTMLMail == nil && req.EmailAddress == "" {
		return errors.New("no request parameter given")
	}
	res := email.Service.Set(c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData, req)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}
