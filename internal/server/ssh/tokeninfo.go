package ssh

import (
	"github.com/gliderlabs/ssh"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

func handleIntrospect(s ssh.Session) error {
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle tokeninfo introspect from ssh")

	var res *model.Response
	var errRes *model.Response
	_ = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			errRes = auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData)
			if errRes != nil {
				return errors.New("dummy")
			}
			res = tokeninfo.HandleTokenInfoIntrospect(rlog, tx, mt, model.ResponseTypeToken, clientMetaData)
			if res.Status >= 400 {
				errRes = res
				return errors.New("dummy")
			}
			return nil
		},
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	return writeJSON(s, res.Response)
}

func handleHistory(s ssh.Session) error {
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle tokeninfo history from ssh")

	var res *model.Response
	var errRes *model.Response
	_ = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			errRes = auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData)
			if errRes != nil {
				return errors.New("dummy")
			}
			res = tokeninfo.HandleTokenInfoHistory(rlog, tx, &pkg.TokenInfoRequest{}, mt, clientMetaData)
			if res.Status >= 400 {
				errRes = res
				return errors.New("dummy")
			}
			return nil
		},
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	return writeJSON(s, res.Response)
}

func handleSubtokens(s ssh.Session) error {
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle tokeninfo subtokens from ssh")

	var res *model.Response
	var errRes *model.Response
	_ = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			errRes = auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
			if errRes != nil {
				return errors.New("dummy")
			}
			res = tokeninfo.HandleTokenInfoSubtokens(rlog, tx, &pkg.TokenInfoRequest{}, mt, clientMetaData)
			if res.Status >= 400 {
				errRes = res
				return errors.New("dummy")
			}
			return nil
		},
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	return writeJSON(s, res.Response)
}

func handleListMytokens(s ssh.Session) error {
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle tokeninfo list mytokens from ssh")

	var res *model.Response
	var errRes *model.Response
	_ = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			errRes = auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
			if errRes != nil {
				return errors.New("dummy")
			}
			res = tokeninfo.HandleTokenInfoList(rlog, tx, &pkg.TokenInfoRequest{}, mt, clientMetaData)
			if res.Status >= 400 {
				errRes = res
				return errors.New("dummy")
			}
			return nil
		},
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	return writeJSON(s, res.Response)
}
