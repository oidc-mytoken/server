package tokeninfo

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/eventrepo"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	eventService "github.com/oidc-mytoken/server/shared/mytoken/event"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/shared/mytoken/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
	"github.com/oidc-mytoken/server/shared/mytoken/rotation"
)

func doTokenInfoHistory(
	rlog log.Ext1FieldLogger, req pkg.TokenInfoRequest, mt *mytoken.Mytoken, clientMetadata *api.ClientMetaData,
	usedRestriction *restrictions.Restriction,
) (history eventrepo.EventHistory, tokenUpdate *response.MytokenResponse, err error) {
	err = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			history, err = eventrepo.GetEventHistory(rlog, tx, mt.ID)
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
					Event: event.FromNumber(event.TokenInfoHistory, ""),
					MTID:  mt.ID,
				}, *clientMetadata,
			)
		},
	)
	return
}

func handleTokenInfoHistory(
	rlog log.Ext1FieldLogger, req pkg.TokenInfoRequest, mt *mytoken.Mytoken, clientMetadata *api.ClientMetaData,
) model.Response {
	// If we call this function it means the token is valid.
	usedRestriction, errRes := auth.CheckCapabilityAndRestriction(
		rlog, nil, mt, clientMetadata.IP, nil, nil, api.CapabilityTokeninfoHistory,
	)
	if errRes != nil {
		return *errRes
	}
	history, tokenUpdate, err := doTokenInfoHistory(rlog, req, mt, clientMetadata, usedRestriction)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return *model.ErrorToInternalServerErrorResponse(err)
	}
	rsp := pkg.NewTokeninfoHistoryResponse(history, tokenUpdate)
	return makeTokenInfoResponse(rsp, tokenUpdate)
}
