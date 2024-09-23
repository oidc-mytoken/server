package notifier

import (
	"encoding/json"
	"time"

	utils2 "github.com/oidc-mytoken/utils/utils"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/notifier/pkg"
	"github.com/oidc-mytoken/server/internal/notifier/server/mailing"
	"github.com/oidc-mytoken/server/internal/utils"
	"github.com/oidc-mytoken/server/internal/utils/cache"
)

// ServerPaths holds the server paths
var ServerPaths = Paths{
	Email: "/email",
}

// Paths holds the server paths
type Paths struct {
	Email string
}

// Prefix prefixes all values in Paths with the specified prefix by url-combining them
func (p Paths) Prefix(prefix string) (out Paths) {
	m := utils.StructToStringMapUsingJSONTags(p)
	for k, v := range m {
		m[k] = utils2.CombineURLPath(prefix, v)
	}
	jsonData, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(jsonData, &out); err != nil {
		panic(err)
	}
	return
}

// InitStandalone initializes a standalone notifier server
func InitStandalone(mailConf config.MailNotificationConf) {
	initCommon(mailConf)
	cache.SetCache(cache.NewInternalCache(3 * time.Minute))
	startServer()
}

// InitIntegrated initializes the integrated notifier "server"
func InitIntegrated() {
	initCommon(config.Get().Features.Notifications.Mail)
}

func initCommon(mailConf config.MailNotificationConf) {
	mailing.Init(mailConf)
	// TODO at this place we would spin up ws
}

// HandleEmailRequest handles a pkg.EmailNotificationRequest
func HandleEmailRequest(req pkg.EmailNotificationRequest) error {
	log.WithField("req", req).Info("Handling email request")
	sID := req.ScheduleID
	returnError := func(err error) error {
		return err
	}
	if sID != "" {
		if found, err := cache.Get(cache.ScheduledNotifications, sID, &struct{}{}); err == nil && found {
			return nil
		} else {
			returnError = func(err error) error {
				if err == nil {
					return cache.Set(cache.ScheduledNotifications, sID, struct{}{})
				}
				return err
			}
		}
	}
	if req.ICSInvite {
		if err := mailing.ICSMailSender.Send(req.To, req.Subject, req.Text, req.Attachments...); err != nil {
			log.WithError(err).Error("error while sending ics mail invite")
			return returnError(err)
		}
		return returnError(nil)
	}
	sender := mailing.PlainTextMailSender
	if req.PreferHTML {
		sender = mailing.HTMLMailSender
	}
	if req.Template != "" {
		if err := sender.SendTemplate(req.To, req.Subject, req.Template, req.BindingData); err != nil {
			log.WithError(err).Error("error while sending templated mail")
			return returnError(err)
		}
		return returnError(nil)
	}
	if err := sender.Send(req.To, req.Subject, req.Text, req.Attachments...); err != nil {
		log.WithError(err).Error("error while sending mail")
		return returnError(err)
	}
	return returnError(nil)
}
