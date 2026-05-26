package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/service/tokeninfo"
)

func handleIntrospect(s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle tokeninfo introspect from ssh")
	res := tokeninfo.Service.Introspect(c.rlog, c.mt, model.ResponseTypeToken, c.clientMetaData)
	if res == nil {
		return writeError(s, errors.New("internal server error"))
	}
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeJSON(s, res.Response)
}

func handleHistory(s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle tokeninfo history from ssh")
	res := tokeninfo.Service.History(c.rlog, c.mt, c.clientMetaData, &pkg.TokenInfoRequest{})
	if res == nil {
		return writeError(s, errors.New("internal server error"))
	}
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeJSON(s, res.Response)
}

func handleSubtokens(s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle tokeninfo subtokens from ssh")
	res := tokeninfo.Service.Subtokens(c.rlog, c.mt, c.clientMetaData, &pkg.TokenInfoRequest{})
	if res == nil {
		return writeError(s, errors.New("internal server error"))
	}
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeJSON(s, res.Response)
}

func handleListMytokens(s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle tokeninfo list mytokens from ssh")
	res := tokeninfo.Service.List(c.rlog, c.mt, c.clientMetaData, &pkg.TokenInfoRequest{})
	if res == nil {
		return writeError(s, errors.New("internal server error"))
	}
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeJSON(s, res.Response)
}

func handleTokenInfoNotifications(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle tokeninfo notifications from ssh")

	var req pkg.TokenInfoRequest
	if len(reqData) > 0 {
		if err := json.Unmarshal(reqData, &req); err != nil {
			return err
		}
	}
	res := tokeninfo.Service.Notifications(c.rlog, c.mt, c.clientMetaData, &req)
	if res == nil {
		return writeError(s, errors.New("internal server error"))
	}
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeJSON(s, res.Response)
}
