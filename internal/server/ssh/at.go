package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/endpoints/token/access"
	"github.com/oidc-mytoken/server/internal/endpoints/token/access/pkg"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/shared/model"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/utils"
)

func handleSSHAT(reqData []byte, s ssh.Session) error {
	ctx := s.Context()
	req := pkg.NewAccessTokenRequest()
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
	rlog.Debug("Handle AT from ssh")
	rlog.Trace("Parsed AT request")

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	usedRestriction, errRes := auth.CheckCapabilityAndRestriction(
		rlog, nil, mt, clientMetaData.IP,
		utils.SplitIgnoreEmpty(req.Scope, " "),
		utils.SplitIgnoreEmpty(req.Audience, " "),
		api.CapabilityAT,
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	provider, errRes := auth.RequireMatchingIssuer(rlog, mt.OIDCIssuer, &req.Issuer)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	res := access.HandleAccessTokenRefresh(rlog, mt, req, clientMetaData, provider, usedRestriction)
	if res.Status >= 400 {
		return writeErrRes(s, errRes)
	}
	tokenRes := res.Response.(pkg.AccessTokenResponse)
	return writeString(s, tokenRes.AccessToken)
}
