package notifier

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/httpclient"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/userrepo"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo"
	"github.com/oidc-mytoken/server/internal/model"
	pkg2 "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/notifier/pkg"
	server "github.com/oidc-mytoken/server/internal/notifier/server"
	"github.com/oidc-mytoken/server/internal/notifier/server/mailing"
	"github.com/oidc-mytoken/server/internal/server/routes"
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
	if !config.Get().Features.Notifications.AnyEnabled {
		return
	}
	if nsURL := config.Get().Features.Notifications.NotifierServer; nsURL != "" {
		initStandalone(nsURL)
	} else {
		initIntegraded()
	}
	initScheduler()
}

func initStandalone(serverURL string) {
	notifier = standaloneNotifier{
		serverAddress: serverURL,
		paths:         server.ServerPaths.Prefix(serverURL),
	}
}
func initIntegraded() {
	notifier = integratedNotifier{}
	server.InitIntegrated()
}

type standaloneNotifier struct {
	serverAddress string
	paths         server.Paths
}
type integratedNotifier struct{}

// SendEmailRequest sends a pkg.EmailNotificationRequest to the standalone notifier server
func (n standaloneNotifier) SendEmailRequest(req pkg.EmailNotificationRequest) {
	_, err := httpclient.Do().R().
		SetBody(req).
		Post(n.paths.Email)
	if err != nil {
		log.WithError(err).Error("error while sending notification request to notifier server")
	}
}

// SendEmailRequest sends a pkg.EmailNotificationRequest to the integrated notification server
func (integratedNotifier) SendEmailRequest(req pkg.EmailNotificationRequest) {
	_ = server.HandleEmailRequest(req)
}

// SendTemplateEmail sends a templated email through the relevant notification server
func SendTemplateEmail(to, subject string, preferHTML bool, template string, binding any) {
	// skipcq GO-E1007
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

// SendICSMail sends an ics calendar invite via email through the relevant notification server
func SendICSMail(to, subject, text string, attachments ...mailing.Attachment) {
	// skipcq GO-E1007
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

// SendNotificationsForEvent sends all relevant notifications for an event through the relevant notification server,
// if there are any
func SendNotificationsForEvent(rlog log.Ext1FieldLogger, tx *sqlx.Tx, e pkg2.MTEvent) error {
	rlog.WithField("event", e.Event.String()).Debug("checking and sending notification for event")
	nc := api.NotificationClassFromEvent(e.Event)
	if nc == nil {
		return nil
	}
	rlog.WithField("notification_class", nc.Name).Trace("found notification class")
	notifications, err := notificationsrepo.GetNotificationsForMTAndClass(rlog, tx, e.MTID, nc)
	if err != nil {
		return err
	}
	if len(notifications) == 0 {
		return nil
	}
	rlog.WithField("number_notifications", len(notifications)).Trace("found notifications for the class and token")
	return sendNotificationsForNotificationInfos(rlog, tx, e.MTID, notifications, nc.Name, &e.ClientMetaData, &e, nil)
}

// SendNotificationsForSubClass sends all relevant notifications for a Notification(
// sub)class through the relevant notification server, if there are any
func SendNotificationsForSubClass(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID,
	nc *api.NotificationClass, clientData *api.ClientMetaData, additionalData model.KeyValues,
	additionalCallbackCheck func() (bool, error),
) error {
	rlog.WithField("notification_class", nc.Name).Debug("checking and sending notification for (sub)class")
	allNotifications, err := notificationsrepo.GetNotificationsForMT(rlog, tx, mtID)
	if err != nil {
		return err
	}
	rlog.WithField("number_all_notifications", len(allNotifications)).Trace("found notifications for token")
	var notifications []api.NotificationInfoBase
	for _, n := range allNotifications {
		thisNC := api.NewNotificationClass(n.Class)
		if thisNC.Contains(nc) {
			notifications = append(notifications, n.NotificationInfoBase.NotificationInfoBase)
		}
	}
	if len(notifications) == 0 {
		return nil
	}
	if additionalCallbackCheck != nil {
		ok, err := additionalCallbackCheck()
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}
	rlog.WithField("number_filtered_notifications", len(notifications)).Trace("found notifications for token and class")
	return sendNotificationsForNotificationInfos(
		rlog, tx, mtID, notifications, nc.Name, clientData, nil, additionalData,
	)
}

func sendNotificationsForNotificationInfos(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID,
	notifications []api.NotificationInfoBase, notificationClassName string,
	clientData *api.ClientMetaData,
	e *pkg2.MTEvent, additionalData model.KeyValues,
) error {
	mailAlreadySent := false
	for _, n := range notifications {
		switch n.Type {
		case api.NotificationTypeMail:
			if !mailAlreadySent {
				mailAlreadySent = true
				emailInfo, err := userrepo.GetMail(rlog, tx, mtID)
				if err != nil {
					return err
				}
				if !emailInfo.Mail.Valid {
					return errors.New("no email set for user")
				}
				if !emailInfo.MailVerified {
					return errors.New("notification email not verified")
				}
				tokenName, err := mytokenrepohelper.GetMTName(rlog, tx, mtID)
				if err != nil {
					return err
				}
				bindingData := map[string]any{
					"management-url": routes.NotificationManagementURL(n.ManagementCode),
				}
				if emailInfo.PreferHTMLMail {
					bindingData["ip"] = clientData.IP
					bindingData["user-agent"] = clientData.UserAgent
					bindingData["country"] = geoip.Country(clientData.IP)
					bindingData["notification-class"] = notificationClassName
					bindingData["mom_id"] = mtID.Hash()
					bindingData["token-name"] = tokenName.String
					if e != nil {
						bindingData["event"] = e.Event.String()
						bindingData["comment"] = e.Comment
					}
					if additionalData != nil {
						bindingData["additional-data"] = additionalData
					}
				} else {
					tableData := map[string]string{}
					if tokenName.Valid {
						tableData["Mytoken Name"] = tokenName.String
					}
					tableData["Mytoken Mom ID"] = mtID.Hash()
					tableData["IP"] = clientData.IP
					tableData["User-Agent"] = clientData.UserAgent
					if country := geoip.Country(clientData.IP); country != "" {
						tableData["Location"] = country
					}
					tableData["Notification Reason"] = notificationClassName

					if e != nil {
						tableData["Event"] = e.Event.String()
						tableData["Comment"] = e.Comment
					}
					for _, kv := range additionalData {
						tableData[kv.Key] = fmt.Sprintf("%v", kv.Value)
					}
					txtTable := generateSimpleTable(nil, tableData)
					bindingData["txt-table"] = txtTable
				}
				rlog.Debug("sending notification mail")
				SendTemplateEmail(
					emailInfo.Mail.String, fmt.Sprintf("mytoken notification: %s", notificationClassName),
					emailInfo.PreferHTMLMail, "notification", bindingData,
				)

			}
		case api.NotificationTypeWebsocket:
			return errors.New("not yet implemented")
		}

	}
	return nil
}
