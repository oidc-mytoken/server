package tokeninfo

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/tokeninfo/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	pkg2 "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

// Introspect returns token introspection information.
func (s *service) Introspect(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken,
	originalTokenType model.ResponseType, clientMetaData *api.ClientMetaData,
) *model.Response {
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if errRes := auth.RequireCapability(
				rlog, tx, api.CapabilityTokeninfoIntrospect, mt, clientMetaData,
			); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			var rollback bool
			res, rollback = s.introspectLogic(rlog, tx, mt, originalTokenType, clientMetaData)
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

func (s *service) introspectLogic(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mt *mytoken.Mytoken,
	originalTokenType model.ResponseType, clientMetaData *api.ClientMetaData,
) (*model.Response, bool) {
	var usedToken mytoken.UsedMytoken
	var tags []api.MTTagInfo

	tmp, err := mt.ToUsedMytoken(rlog, tx)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err), true
	}
	usedToken = *tmp

	tags, err = mytokenrepo.GetTags(rlog, tx, mt.ID)
	if err != nil {
		rlog.WithError(err).Debug("could not get tags for token")
		tags = []api.MTTagInfo{}
	}

	err = eventService.LogEvent(
		rlog, tx, pkg2.MTEvent{
			Event:          api.EventTokenInfoIntrospect,
			MTID:           mt.ID,
			ClientMetaData: *clientMetaData,
		},
	)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err), true
	}

	return &model.Response{
		Status: fiber.StatusOK,
		Response: pkg.TokeninfoIntrospectResponse{
			TokeninfoIntrospectResponse: api.TokeninfoIntrospectResponse{
				Valid: true,
				MOMID: mt.ID.Hash(),
			},
			Token:     usedToken,
			TokenType: originalTokenType,
			Tags:      tags,
		},
	}, false
}
