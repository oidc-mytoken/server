package ssh

import (
	"encoding/json"

	"github.com/gliderlabs/ssh"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/userrepo"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	"github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

func handleSSHEmailGet(s ssh.Session) error {
	if !config.Get().Features.Notifications.Mail.Enabled {
		return errors.New("mail notifications are disabled")
	}
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle email-get from ssh")

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	_, errRes = auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt, clientMetaData, api.CapabilityEmailRead,
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}

	var info userrepo.MailInfo
	var res *model.Response
	_ = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			var err error
			info, err = userrepo.GetMail(rlog, tx, mt.ID)
			if err != nil {
				res = model.ErrorToInternalServerErrorResponse(err)
				return err
			}
			return nil
		},
	)
	if res != nil {
		return writeErrRes(s, res)
	}
	return writeJSON(
		s, api.MailSettingsInfoResponse{
			EmailAddress:   info.Mail.String,
			EmailVerified:  info.MailVerified,
			PreferHTMLMail: info.PreferHTMLMail,
		},
	)
}

func handleSSHEmailSet(reqData []byte, s ssh.Session) error {
	if !config.Get().Features.Notifications.Mail.Enabled {
		return errors.New("mail notifications are disabled")
	}
	ctx := s.Context()
	mt := ctx.Value("mytoken").(*mytoken.Mytoken)
	clientMetaData := &api.ClientMetaData{
		IP:        ctx.Value("ip").(string),
		UserAgent: ctx.Value("user_agent").(string),
	}
	rlog := logger.GetSSHRequestLogger(ctx.Value("session").(string))
	rlog.Debug("Handle email-set from ssh")

	var req api.UpdateMailSettingsRequest
	if err := json.Unmarshal(reqData, &req); err != nil {
		return err
	}
	if req.PreferHTMLMail == nil && req.EmailAddress == "" {
		return errors.New("no request parameter given")
	}

	errRes := auth.RequireMytokenNotRevoked(rlog, nil, mt, clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt, clientMetaData, api.CapabilityEmail,
	)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}

	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if req.PreferHTMLMail != nil {
				if err := userrepo.ChangePreferredMailType(rlog, tx, mt.ID, *req.PreferHTMLMail); err != nil {
					return err
				}
				eventComment := "to plain text"
				if *req.PreferHTMLMail {
					eventComment = "to html"
				}
				if err := eventService.LogEvent(
					rlog, tx, pkg.MTEvent{
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
				if err := eventService.LogEvent(
					rlog, tx, pkg.MTEvent{
						Event:          api.EventEmailChanged,
						Comment:        req.EmailAddress,
						MTID:           mt.ID,
						ClientMetaData: *clientMetaData,
					},
				); err != nil {
					return err
				}
			}
			if err := usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	return writeString(s, "OK")
}
