package notifier

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/unixtime"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/userrepo"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/actions"
	"github.com/oidc-mytoken/server/internal/server/routes"
)

func initScheduler() {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			checkingForDueNotifications()
		}
	}()
}

func checkingForDueNotifications() {
	logger := log.StandardLogger()
	logger.Trace("Checking for notifications to send")
	var abort bool
	for {
		if err := db.Transact(
			logger, func(tx *sqlx.Tx) error {
				n, err := notificationsrepo.PopOneScheduledNotification(logger, tx)
				if err != nil {
					log.WithError(err).Error("error popping scheduled notification")
					return err
				}
				if n == nil {
					abort = true
					return nil
				}
				return handleDueNotification(logger, tx, n)
			},
		); err != nil {
			break
		}
		if abort {
			logger.Trace("No notifications due")
			break
		}
	}
}

func handleDueNotification(logger log.Ext1FieldLogger, tx *sqlx.Tx, n *notificationsrepo.ScheduledNotification) error {
	logger.Trace("Got a notification")
	switch n.Type {
	case api.NotificationTypeMail:
		return handleDueMailNotification(logger, tx, n)
	case api.NotificationTypeWebsocket:
		// NYI
		return nil
	default:
		return errors.New("unknown notification type")
	}
}

func handleDueMailNotification(
	logger log.Ext1FieldLogger, tx *sqlx.Tx,
	n *notificationsrepo.ScheduledNotification,
) error {
	emailInfo, err := userrepo.GetMail(logger, tx, n.MTID)
	if err != nil {
		return err
	}
	if !emailInfo.MailVerified {
		return nil
	}
	name, err := mytokenrepohelper.GetMTName(logger, tx, n.MTID)
	if err != nil {
		return err
	}
	var subject string
	var template string
	var bindingData map[string]any
	switch n.Class {
	case notificationsrepo.ScheduleClassExp:
		exp_, ok := n.AdditionalInfo[notificationsrepo.AdditionalInfoKeyExpiresAt].(float64)
		if !ok {
			logger.Error("'expires_at' missing or wrong time in scheduled notification of class 'exp'")
			return nil
		}
		exp := unixtime.UnixTime(exp_)
		diff := time.Until(exp.Time())
		var diffStr string
		if diff < 24*time.Hour {
			diff = diff.Round(time.Hour)
			diffStr = fmt.Sprintf("%d hours", diff/time.Hour)
		} else {
			diff = diff.Round(24 * time.Hour)
			diffStr = fmt.Sprintf("%d days", diff/(24*time.Hour))
		}
		var quotedName string
		if name.Valid {
			quotedName = fmt.Sprintf(" '%s'", name.String)
		}
		subject = fmt.Sprintf("mytoken%s expires in %s", quotedName, diffStr)
		template = "notification-exp"
		recreateURL, err := actions.CreateRecreateToken(logger, tx, n.MTID)
		if err != nil {
			return err
		}
		unsubscribeURL, err := actions.GetUnsubscribeScheduled(
			logger, tx, n.MTID, n.NotificationID,
		)
		if err != nil {
			return err
		}
		bindingData = map[string]any{
			"expires_at":                     exp.Time().String(),
			"token-name":                     name,
			"mom_id":                         n.MTID.Hash(),
			"management-url":                 routes.NotificationManagementURL(n.ManagementCode),
			"recreate-url":                   recreateURL,
			"unsubscribe-exp-this-token-url": unsubscribeURL,
		}
	default:
		return nil
	}
	SendTemplateEmail(
		emailInfo.Mail.String, subject, emailInfo.PreferHTMLMail, template, bindingData,
	)
	return nil
}
