package tokeninfo

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	event "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/rotation"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

func doTokenInfoList(
	rlog log.Ext1FieldLogger, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken, clientMetadata *api.ClientMetaData,
	usedRestriction *restrictions.Restriction,
) (tokenList []*tree.MytokenEntryTree, tokenUpdate *response.MytokenResponse, err error) {
	err = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			tokenList, err = tree.AllTokens(rlog, tx, mt.ID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			if usedRestriction == nil {
				return nil
			}
			if err = usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
				return err
			}
			tokenUpdate, err = rotation.RotateMytokenAfterOtherForResponse(
				rlog, tx, req.Mytoken.JWT, mt, *clientMetadata, req.Mytoken.OriginalTokenType,
			)
			if err != nil {
				return err
			}
			return eventService.LogEvent(
				rlog, tx, eventService.MTEvent{
					Event: event.FromNumber(event.TokenInfoListMTs, ""),
					MTID:  mt.ID,
				}, *clientMetadata,
			)
		},
	)
	return
}

// HandleTokenInfoList handles a tokeninfo list request
func HandleTokenInfoList(
	rlog log.Ext1FieldLogger, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken, clientMetadata *api.ClientMetaData,
) model.Response {
	// If we call this function it means the token is valid.
	usedRestriction, errRes := auth.CheckCapabilityAndRestriction(
		rlog, nil, mt, clientMetadata.IP, nil, nil, api.CapabilityListMT,
	)
	if errRes != nil {
		return *errRes
	}
	tokenList, tokenUpdate, err := doTokenInfoList(rlog, req, mt, clientMetadata, usedRestriction)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return *model.ErrorToInternalServerErrorResponse(err)
	}
	rsp := pkg.NewTokeninfoListResponse(tokenList, tokenUpdate)
	return makeTokenInfoResponse(rsp, tokenUpdate)

}
