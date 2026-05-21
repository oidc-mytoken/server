package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	helper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/model"
	mytokenPkg "github.com/oidc-mytoken/server/internal/mytoken"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

func handleSSHRevoke(reqData []byte, s ssh.Session) error {
	if !config.Get().Features.TokenRevocation.Enabled {
		return errors.New("revocation is disabled")
	}
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle revoke from ssh")

	if len(reqData) == 0 {
		errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
		if errRes != nil {
			return writeErrRes(s, errRes)
		}
		umt := mt.ToUniversalMytoken()
		res := mytokenPkg.RevokeMytoken(rlog, nil, mt.ID, umt.JWT, false, mt.OIDCIssuer)
		if res != nil {
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
	return handleSSHRevokeByMOMID(rlog, s, mt, clientMetaData, req)
}

func handleSSHRevokeByMOMID(
	rlog log.Ext1FieldLogger, s ssh.Session, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData, req api.RevocationRequest,
) error {
	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	var res *model.Response
	_ = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			isParent, err := helper.MOMIDHasParent(rlog, nil, req.MOMID, mt.ID)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if !isParent && !mt.Capabilities.Has(api.CapabilityRevokeAnyToken) {
				res = &model.Response{
					Status: 403,
					Response: api.Error{
						Error:            api.ErrorStrInsufficientCapabilities,
						ErrorDescription: "The provided token is neither a parent of the token to be revoked nor does it have the 'revoke_any_token' capability",
					},
				}
				return errors.New("rollback")
			}
			same, err := helper.CheckMytokensAreForSameUser(rlog, nil, req.MOMID, mt.ID)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			if !same {
				res = &model.Response{
					Status: 403,
					Response: api.Error{
						Error:            api.ErrorStrInvalidGrant,
						ErrorDescription: "The provided token cannot be used to revoke this mom_id",
					},
				}
				return errors.New("rollback")
			}
			if req.MOMID == mt.ID.Hash() {
				res = &model.Response{
					Status: 400,
					Response: api.Error{
						Error:            api.ErrorStrInvalidRequest,
						ErrorDescription: "A token cannot be revoked by its own mom_id. Use the token itself instead.",
					},
				}
				return errors.New("rollback")
			}
			if err = helper.RevokeMT(rlog, tx, req.MOMID, req.Recursive); err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			return nil
		},
	)
	if res != nil {
		return writeErrRes(s, res)
	}
	return writeString(s, "OK")
}
