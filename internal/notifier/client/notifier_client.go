package notifier

import (
	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/notifier/pkg"
	server "github.com/oidc-mytoken/server/internal/notifier/server"
	"github.com/oidc-mytoken/server/internal/notifier/server/mailing"
)

type notifierClient interface {
	SendEmailRequest(req pkg.EmailNotificationRequest)
}

var notifier notifierClient

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

func (n standaloneNotifier) SendEmailRequest(req pkg.EmailNotificationRequest) {
	// todo
}

func (n integratedNotifier) SendEmailRequest(req pkg.EmailNotificationRequest) {
	server.HandleEmailRequest(req)
}

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
