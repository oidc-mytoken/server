package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/shared/model"
	mytoken2 "github.com/oidc-mytoken/server/shared/mytoken"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/utils/ternary"
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
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	req.Mytoken = mt.ToUniversalMytoken()
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle mytoken from ssh")

	req.Restrictions.ReplaceThisIp(clientMetaData.IP)
	req.Restrictions.ClearUnsupportedKeys()
	rlog.Trace("Parsed mytoken request")

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	usedRestriction, errRes := auth.RequireUsableRestriction(
		rlog, nil, mt, clientMetaData.IP, nil, nil, api.CapabilityCreateMT,
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	if _, errRes = auth.RequireMatchingIssuer(rlog, mt.OIDCIssuer, &req.Issuer); errRes != nil {
		return writeErrRes(s, errRes)
	}
	res := mytoken2.HandleMytokenFromMytokenReq(rlog, mt, req, &clientMetaData, usedRestriction)
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	tokenRes := res.Response.(pkg.MytokenResponse)
	return writeString(s, ternary.IfNotEmptyOr(tokenRes.Mytoken, tokenRes.TransferCode))
}
