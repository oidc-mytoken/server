package notifier

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/userrepo"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo"
	pkg2 "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	"github.com/oidc-mytoken/server/internal/notifier/pkg"
	server "github.com/oidc-mytoken/server/internal/notifier/server"
	"github.com/oidc-mytoken/server/internal/notifier/server/mailing"
	"github.com/oidc-mytoken/server/internal/utils/geoip"
)

type notifierClient interface {
	SendEmailRequest(req pkg.EmailNotificationRequest)
}

var notifier notifierClient

// Init initializes the notifier client,
// depending on the server deployment model either for communicating with a standalone notifier server or with the
// integrated one.
func Init() {
	if config.Get().Server.DistributedServers {
		initStandalone()
	} else {
		initIntegraded()
	}
}

func initStandalone() {
	//TODO
	notifier = standaloneNotifier{}
}
func initIntegraded() {
	//TODO
	notifier = integratedNotifier{}
	server.InitIntegraded()
}

type standaloneNotifier struct{}
type integratedNotifier struct{}

// SendEmailRequest sends a pkg.EmailNotificationRequest to the standalone notifier server
func (n standaloneNotifier) SendEmailRequest(req pkg.EmailNotificationRequest) {
	// todo
}

// SendEmailRequest sends a pkg.EmailNotificationRequest to the integrated notification server
func (n integratedNotifier) SendEmailRequest(req pkg.EmailNotificationRequest) {
	server.HandleEmailRequest(req)
}

// SendTemplateEmail sends a templated email through the relevant notification server
func SendTemplateEmail(to, subject string, preferHTML bool, template string, binding any) {
	go func() {
		req := pkg.EmailNotificationRequest{
			To:          to,
			Subject:     subject,
			PreferHTML:  preferHTML,
			Template:    template,
			BindingData: binding,
		}
		notifier.SendEmailRequest(req)
	}()
}

// SendICSMail sends a ics calendar invite via email through the relevant notification server
func SendICSMail(to, subject, text string, attachments ...mailing.Attachment) {
	go func() {
		req := pkg.EmailNotificationRequest{
			To:          to,
			Subject:     subject,
			Text:        text,
			Attachments: attachments,
			ICSInvite:   true,
		}
		notifier.SendEmailRequest(req)
	}()
}

// SendNotifcationsForEvent sends all relevant notifications for a event through the relevant notification server,
// if there are any
func SendNotificationsForEvent(rlog log.Ext1FieldLogger, tx *sqlx.Tx, e pkg2.MTEvent) error {
	rlog.WithField("event", e.Event.String()).Debug("checking and sending notification for event")
	nc := api.NotificationClassFromEvent(e.Event)
	if nc == nil {
		return nil
	}
	rlog.WithField("notification_class", nc.Name).Trace("found notification class")
	notifications, err := notificationsrepo.GetNotificationsForMTAndClass(rlog, tx, e.MTID, *nc)
	if err != nil {
		return err
	}
	rlog.WithField("number_notifications", len(notifications)).Trace("found notifications for the class and token")
	mailAlreadySent := false
	for _, n := range notifications {
		switch n.Type {
		case api.NotificationTypeMail:
			if !mailAlreadySent {
				mailAlreadySent = true
				emailInfo, err := userrepo.GetMail(rlog, tx, e.MTID)
				if err != nil {
					return err
				}
				if !emailInfo.MailVerified {
					return errors.New("notification email not verified")
				}
				tokenName, err := mytokenrepohelper.GetMTName(rlog, tx, e.MTID)
				if err != nil {
					return err
				}
				bindingData := map[string]any{
					"ip":                 e.IP,
					"user-agent":         e.UserAgent,
					"country":            geoip.Country(e.IP),
					"notification-class": nc.Name,
					"event":              e.Event.String(),
					"mom_id":             e.MTID.Hash(),
					"token-name":         tokenName.String,
					"comment":            e.Comment,
					"management-url":     n.ManagementCode, //TODO
				}
				rlog.Debug("sending notification mail")
				SendTemplateEmail(
					emailInfo.Mail, fmt.Sprintf("mytoken notification: %s", nc.Name),
					emailInfo.PreferHTMLMail, "notification", bindingData,
				)

			}
		case api.NotificationTypeWebsocket:
			return errors.New("not yet implemented")
		}

	}
	return nil
}
