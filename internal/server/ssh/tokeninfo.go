package ssh

import (
	"github.com/gliderlabs/ssh"
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/logger"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
)

func handleIntrospect(s ssh.Session) error {
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle tokeninfo introspect from ssh")

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	res := tokeninfo.HandleTokenInfoIntrospect(rlog, mt, &clientMetaData)
	if res.Status >= 400 {
		return writeErrRes(s, &res)
	}
	return writeJSON(s, res.Response)
}

func handleHistory(s ssh.Session) error {
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle tokeninfo history from ssh")

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	res := tokeninfo.HandleTokenInfoHistory(rlog, pkg.TokenInfoRequest{}, mt, &clientMetaData)
	if res.Status >= 400 {
		return writeErrRes(s, &res)
	}
	return writeJSON(s, res.Response)
}

func handleSubtokens(s ssh.Session) error {
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle tokeninfo subtokens from ssh")

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	res := tokeninfo.HandleTokenInfoSubtokens(rlog, pkg.TokenInfoRequest{}, mt, &clientMetaData)
	if res.Status >= 400 {
		return writeErrRes(s, &res)
	}
	return writeJSON(s, res.Response)
}

func handleListMytokens(s ssh.Session) error {
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle tokeninfo list mytokens from ssh")

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	res := tokeninfo.HandleTokenInfoList(rlog, pkg.TokenInfoRequest{}, mt, &clientMetaData)
	if res.Status >= 400 {
		return writeErrRes(s, &res)
	}
	return writeJSON(s, res.Response)
}
