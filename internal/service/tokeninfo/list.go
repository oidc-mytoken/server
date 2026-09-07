package tokeninfo

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	pkg2 "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/rotation"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

// List returns a list of mytokens for the user.
func (s *service) List(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData, req *pkg.TokenInfoRequest,
) *model.Response {
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			var rollback bool
			res, rollback = s.listLogic(rlog, tx, req, mt, clientMetaData)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil {
		if res != nil {
			return res
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return res
}

func (s *service) listLogic(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData,
) (*model.Response, bool) {
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, tx, mt, clientMetaData, api.CapabilityListMT,
	)
	if errRes != nil {
		return errRes, true
	}
	tokenList, tokenUpdate, err := s.doTokenInfoList(rlog, tx, req, mt, clientMetaData, usedRestriction)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err), true
	}
	rsp := pkg.NewTokeninfoListResponse(tokenList, tokenUpdate)
	return s.makeTokenInfoResponse(rsp, tokenUpdate), false
}

func (s *service) doTokenInfoList(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken,
	clientMetaData *api.ClientMetaData,
	usedRestriction *restrictions.Restriction,
) (tokenList []*tree.MytokenEntryTree, tokenUpdate *response.MytokenResponse, err error) {
	tokenList, err = tree.AllTokens(rlog, tx, mt.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}
	if usedRestriction == nil {
		return
	}
	if err = usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
		return
	}
	tokenUpdate, err = rotation.RotateMytokenAfterOtherForResponse(
		rlog, tx, req.Mytoken.JWT, mt, *clientMetaData, req.Mytoken.OriginalTokenType,
	)
	if err != nil {
		return
	}
	err = eventService.LogEvent(
		rlog, tx, pkg2.MTEvent{
			Event:          api.EventTokenInfoListMTs,
			MTID:           mt.ID,
			ClientMetaData: *clientMetaData,
		},
	)
	return
}
