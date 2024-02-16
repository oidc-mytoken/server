package notifier

import (
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/notifier/pkg"
	"github.com/oidc-mytoken/server/internal/notifier/server/mailing"
)

func InitStandalone(mailConf config.MailNotificationConf) {
	initCommon(mailConf)
	startServer()
}
func InitIntegraded() {
	initCommon(config.Get().Features.Notifications.Mail)
}

func initCommon(mailConf config.MailNotificationConf) {
	mailing.Init(mailConf)
	//TODO spin up ws
}

func HandleEmailRequest(req pkg.EmailNotificationRequest) error {
	log.WithField("req", req).Info("Handling email request")
	if req.ICSInvite {
		if err := mailing.ICSMailSender.Send(req.To, req.Subject, req.Text, req.Attachments...); err != nil {
			log.WithError(err).Error("error while sending ics mail invite")
			return err
		}
		return nil
	}
	sender := mailing.PlainTextMailSender
	if req.PreferHTML {
		sender = mailing.HTMLMailSender
	}
	if req.Template != "" {
		if err := sender.SendTemplate(req.To, req.Subject, req.Template, req.BindingData); err != nil {
			log.WithError(err).Error("error while sending templated mail")
			return err
		}
		return nil
	}
	if err := sender.Send(req.To, req.Subject, req.Text, req.Attachments...); err != nil {
		log.WithError(err).Error("error while sending mail")
		return err
	}
	return nil
}
