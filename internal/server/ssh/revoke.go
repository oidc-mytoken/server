package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/service/revoke"
)

func handleSSHRevoke(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	c.rlog.Debug("Handle revoke from ssh")

	if len(reqData) == 0 {
		res := revoke.Service.RevokeSelf(c.rlog, c.mt, c.mt.ToUniversalMytoken(), c.clientMetaData)
		if res != nil && res.Status >= 400 {
			return writeErrRes(s, res)
		}
		return writeString(s, "OK")
	}

	var req api.RevocationRequest
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.MOMID == "" {
		return errors.New("mom_id is required")
	}
	res := revoke.Service.ByMOMID(c.rlog, nil, c.mt, c.clientMetaData, req.MOMID, req.Recursive)
	if res != nil && res.Status >= 400 {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}
