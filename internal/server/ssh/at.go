package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/endpoints/token/access/pkg"
	"github.com/oidc-mytoken/server/internal/service/access"
)

func handleSSHAT(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	req := pkg.NewAccessTokenRequest()
	if len(reqData) > 0 {
		if err := json.Unmarshal(reqData, &req); err != nil {
			if err.Error() != "token not valid" {
				return err
			}
		}
	}
	c.rlog.Debug("Handle AT from ssh")
	req.Mytoken = c.mt.ToUniversalMytoken()
	res := access.Service.CreateAccessToken(c.rlog, c.mt, c.clientMetaData, req)
	if res == nil {
		return writeError(s, errors.New("internal server error"))
	}
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	tokenRes := res.Response.(pkg.AccessTokenResponse)
	return writeString(s, tokenRes.AccessToken)
}
