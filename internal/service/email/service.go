// Package email provides a service layer for email settings operations.
package email

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/userrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/actions"
	response "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	pkg2 "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	notifier "github.com/oidc-mytoken/server/internal/notifier/client"
	"github.com/oidc-mytoken/server/internal/notifier/server/mailing/mailtemplates"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/mytokenutils"
)

// mailSettingsResponse wraps the email settings response to support token updates.
type mailSettingsResponse struct {
	api.MailSettingsInfoResponse
	TokenUpdate *response.MytokenResponse `json:"token_update,omitempty"`
}

func (r *mailSettingsResponse) SetTokenUpdate(tu *response.MytokenResponse) {
	r.TokenUpdate = tu
}

// Service is the email service singleton.
var Service = &service{}

type service struct{}

// Get returns the email settings for the user.
func (s *service) Get(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData,
) *model.Response {
	if !config.Get().Features.Notifications.Mail.Enabled {
		return model.BadRequestErrorResponse("mail notifications are disabled")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, clientMetaData, api.CapabilityEmailRead,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			info, dbErr := userrepo.GetMail(rlog, tx, mt.ID)
			if dbErr != nil {
				return dbErr
			}
			res = &model.Response{
				Status: http.StatusOK,
				Response: &mailSettingsResponse{
					MailSettingsInfoResponse: api.MailSettingsInfoResponse{
						EmailAddress:   info.Mail.String,
						EmailVerified:  info.MailVerified,
						PreferHTMLMail: info.PreferHTMLMail,
					},
				},
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, res, mt, *clientMetaData,
				api.EventEmailSettingsListed, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
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

// Set updates the email settings for the user.
func (s *service) Set(
	rlog log.Ext1FieldLogger, mt *mytoken.Mytoken, umt universalmytoken.UniversalMytoken,
	clientMetaData *api.ClientMetaData, req api.UpdateMailSettingsRequest,
) *model.Response {
	if !config.Get().Features.Notifications.Mail.Enabled {
		return model.BadRequestErrorResponse("mail notifications are disabled")
	}
	if req.PreferHTMLMail == nil && req.EmailAddress == "" {
		return model.BadRequestErrorResponse("no request parameter given")
	}
	var res *model.Response
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if errRes := auth.RequireMytokenNotRevoked(rlog, tx, mt, clientMetaData); errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, clientMetaData, api.CapabilityEmail,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			if req.PreferHTMLMail != nil {
				if err := userrepo.ChangePreferredMailType(rlog, tx, mt.ID, *req.PreferHTMLMail); err != nil {
					return err
				}
				eventComment := "to plain text"
				if *req.PreferHTMLMail {
					eventComment = "to html"
				}
				if err := eventService.LogEvent(
					rlog, tx, pkg2.MTEvent{
						Event:          api.EventEmailMimetypeChanged,
						Comment:        eventComment,
						MTID:           mt.ID,
						ClientMetaData: *clientMetaData,
					},
				); err != nil {
					return err
				}
			}
			if req.EmailAddress != "" {
				if err := userrepo.ChangeEmail(rlog, tx, mt.ID, req.EmailAddress); err != nil {
					return err
				}
				verificationURL, err := actions.CreateVerifyEmail(rlog, tx, mt.ID)
				if err != nil {
					return err
				}
				mailInfo, err := userrepo.GetMail(rlog, tx, mt.ID)
				if err != nil {
					return err
				}
				if err := eventService.LogEvent(
					rlog, tx, pkg2.MTEvent{
						Event:          api.EventEmailChanged,
						Comment:        req.EmailAddress,
						MTID:           mt.ID,
						ClientMetaData: *clientMetaData,
					},
				); err != nil {
					return err
				}
				notifier.SendTemplateEmail(
					req.EmailAddress, mailtemplates.SubjectVerifyMail, mailInfo.PreferHTMLMail,
					mailtemplates.TemplateVerifyMail, map[string]any{
						"issuer": config.Get().IssuerURL,
						"link":   verificationURL,
					},
				)
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, nil, mt, *clientMetaData,
				api.EventUnknown, "", usedRestriction, umt.JWT, umt.OriginalTokenType,
			)
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
