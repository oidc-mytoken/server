package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/oidc-mytoken/utils/utils/ternary"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/service/mytoken"
)

func handleSSHMytoken(reqData []byte, s ssh.Session) error {
	c := newSSHSessionCtx(s)
	req := pkg.NewMytokenRequest()
	if len(reqData) > 0 {
		if err := json.Unmarshal(reqData, &req); err != nil {
			return err
		}
	}
	c.rlog.Debug("Handle mytoken from ssh")
	umt := c.mt.ToUniversalMytoken()
	req.Mytoken = umt
	res := mytoken.Service.CreateFromMytoken(c.rlog, c.mt, umt, c.clientMetaData, req)
	if res == nil {
		return writeError(s, errors.New("internal server error"))
	}
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	tokenRes := res.Response.(pkg.MytokenResponse)
	return writeString(s, ternary.IfNotEmptyOr(tokenRes.Mytoken, tokenRes.TransferCode))
}
