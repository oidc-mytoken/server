package mailing

import (
	"fmt"
	"io"
	"net/smtp"
	"time"

	"github.com/jordan-wright/email"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/mailing/mailtemplates"
)

var mailPool *email.Pool
var fromAddress string

// Init initializes the mailing
func Init() {
	if !config.Get().Features.Notifications.Mail.Enabled {
		HTMLMailSender = noopSender{}
		PlainTextMailSender = noopSender{}
		return
	}
	mailServerConfig :=
		config.Get().Features.Notifications.Mail.MailServer
	fromAddress = mailServerConfig.FromAddress
	var err error
	mailPool, err = email.NewPool(
		fmt.Sprintf("%s:%d", mailServerConfig.Host, mailServerConfig.Port),
		4,
		smtp.PlainAuth("", mailServerConfig.Username, mailServerConfig.Password, mailServerConfig.Host),
	)
	if err != nil {
		log.WithError(err).Fatal("could not connect to email server")
	}
	mailtemplates.Init()
}

// SendEMail send the passed email.Email
func SendEMail(mail *email.Email) error {
	return errors.WithStack(mailPool.Send(mail, 10*time.Second))
}

// Attachment is a type holding information about an email attachment
type Attachment struct {
	Reader      io.Reader
	Filename    string
	ContentType string
}

// MailSender is an interface for types that can send mails
type MailSender interface {
	Send(to, subject, text string, attachments ...Attachment) error
}

// TemplateMailSender is an interface for types that can send template mails
type TemplateMailSender interface {
	SendTemplate(to, subject, template string, binding any) error
	MailSender
}

type plainTextMailSender struct{}
type htmlMailSender struct{}
type icsMailSender struct{}
type noopSender struct{}

// PlainTextMailSender is a MailSender that sends plain text mails
var PlainTextMailSender TemplateMailSender = plainTextMailSender{}

// HTMLMailSender is a MailSender that sends html mails
var HTMLMailSender TemplateMailSender = htmlMailSender{}

// ICSMailSender is a MailSender that sends calendar invitations
var ICSMailSender MailSender = icsMailSender{}

// Send implements the MailSender interface
func (s noopSender) Send(_, _, _ string, _ ...Attachment) error {
	return nil
}

// SendTemplate implements the TemplateMailSender interface
func (s noopSender) SendTemplate(_, _, _ string, _ any) error {
	return nil
}

// Send implements the MailSender interface
func (s plainTextMailSender) Send(to, subject, text string, attachments ...Attachment) error {
	mail := &email.Email{
		From:    fromAddress,
		To:      []string{to},
		Subject: subject,
		Text:    []byte(text),
	}
	for _, a := range attachments {
		_, err := mail.Attach(a.Reader, a.Filename, a.ContentType)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return SendEMail(mail)
}

// SendTemplate implements the TemplateMailSender interface
func (s plainTextMailSender) SendTemplate(to, subject, template string, binding any) error {
	text, err := mailtemplates.Text(template, binding)
	if err != nil {
		return err
	}
	return s.Send(to, subject, text)
}

// Send implements the MailSender interface
func (s htmlMailSender) Send(to, subject, text string, attachments ...Attachment) error {
	mail := &email.Email{
		From:    fromAddress,
		To:      []string{to},
		Subject: subject,
		HTML:    []byte(text),
	}
	for _, a := range attachments {
		_, err := mail.Attach(a.Reader, a.Filename, a.ContentType)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return SendEMail(mail)
}

// SendTemplate implements the TemplateMailSender interface
func (s htmlMailSender) SendTemplate(to, subject, template string, binding any) error {
	text, err := mailtemplates.HTML(template, binding)
	if err != nil {
		return err
	}
	return s.Send(to, subject, text)
}

// Send implements the MailSender interface
func (s icsMailSender) Send(to, subject, text string, attachments ...Attachment) error {
	mail := &email.Email{
		From:    fromAddress,
		To:      []string{to},
		Subject: subject,
		Text:    []byte(text),
	}
	for _, a := range attachments {
		aa, err := mail.Attach(a.Reader, a.Filename, a.ContentType)
		if err != nil {
			return errors.WithStack(err)
		}
		aa.Header.Set("Content-Disposition", "inline")
	}
	return SendEMail(mail)
}
