package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils/ternary"

	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken2 "github.com/oidc-mytoken/server/internal/mytoken"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

func handleSSHMytoken(reqData []byte, s ssh.Session) error {
	ctx := s.Context()
	req := pkg.NewMytokenRequest()
	req.GrantType = model.GrantTypeMytoken
	if len(reqData) > 0 {
		if err := json.Unmarshal(reqData, &req); err != nil {
			return err
		}
	}
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	req.Mytoken = ctx.Value("mytoken").(*mytoken.Mytoken).ToUniversalMytoken()
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle mytoken from ssh")

	usedRestriction, mt, errRes := mytoken2.HandleMytokenFromMytokenReqChecks(rlog, req, clientMetaData, nil)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	res := mytoken2.HandleMytokenFromMytokenReq(rlog, mt, req, clientMetaData, usedRestriction)
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	tokenRes := res.Response.(pkg.MytokenResponse)
	return writeString(s, ternary.IfNotEmptyOr(tokenRes.Mytoken, tokenRes.TransferCode))
}
