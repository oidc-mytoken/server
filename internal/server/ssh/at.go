package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/endpoints/token/access"
	"github.com/oidc-mytoken/server/internal/endpoints/token/access/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/utils"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

func handleSSHAT(reqData []byte, s ssh.Session) error {
	ctx := s.Context()
	req := pkg.NewAccessTokenRequest()
	if len(reqData) > 0 {
		if err := json.Unmarshal(reqData, &req); err != nil {
			if err.Error() != "token not valid" {
				return err
			}
		}
	}
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	req.GrantType = model.GrantTypeMytoken
	req.Mytoken = mt.ToUniversalMytoken()
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle AT from ssh")
	rlog.Trace("Parsed AT request")

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestriction(
		rlog, nil, mt, clientMetaData,
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
	res := access.HandleAccessTokenRefresh(rlog, mt, req, *clientMetaData, provider, usedRestriction)
	if res.Status >= 400 {
		return writeErrRes(s, res)
	}
	tokenRes := res.Response.(pkg.AccessTokenResponse)
	return writeString(s, tokenRes.AccessToken)
}
