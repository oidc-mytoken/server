package pkg

import (
	"github.com/oidc-mytoken/server/internal/notifier/server/mailing"
)

// EmailNotificationRequest holds a request to send an email notification
type EmailNotificationRequest struct {
	To          string               `json:"to"`
	Subject     string               `json:"subject"`
	Text        string               `json:"text,omitempty"`
	PreferHTML  bool                 `json:"prefer_html"`
	Template    string               `json:"template,omitempty"`
	BindingData any                  `json:"binding_data,omitempty"`
	ICSInvite   bool                 `json:"ics_invite,omitempty"`
	Attachments []mailing.Attachment `json:"attachments,omitempty"`
	ScheduleID  string               `json:"schedule_id,omitempty"`
}
