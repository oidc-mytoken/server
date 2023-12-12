package tokeninfo

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
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

func doTokenInfoSubtokens(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken,
	clientMetadata *api.ClientMetaData,
	usedRestriction *restrictions.Restriction,
) (tokenTree tree.MytokenEntryTree, tokenUpdate *response.MytokenResponse, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			tokenTree, err = tree.TokenSubTree(rlog, tx, mt.ID)
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
				rlog, tx, pkg2.MTEvent{
					Event:          api.EventTokenInfoSubtokens,
					MTID:           mt.ID,
					ClientMetaData: *clientMetadata,
				},
			)
		},
	)
	return
}

// HandleTokenInfoSubtokens handles a tokeninfo list subtokens request
func HandleTokenInfoSubtokens(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req *pkg.TokenInfoRequest, mt *mytoken.Mytoken,
	clientMetadata *api.ClientMetaData,
) model.Response {
	// If we call this function it means the token is valid.
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, tx, mt, clientMetadata, api.CapabilityTokeninfoSubtokens,
	)
	if errRes != nil {
		return *errRes
	}
	tokenTree, tokenUpdate, err := doTokenInfoSubtokens(rlog, tx, req, mt, clientMetadata, usedRestriction)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return *model.ErrorToInternalServerErrorResponse(err)
	}
	rsp := pkg.NewTokeninfoSubtokensResponse(tokenTree, tokenUpdate)
	return makeTokenInfoResponse(rsp, tokenUpdate)
}
