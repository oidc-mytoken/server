package mailing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/smtp"
	"strings"
	"time"

	"github.com/jordan-wright/email"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/notifier/server/mailing/mailtemplates"
	"github.com/oidc-mytoken/server/internal/utils/multilimiter"
)

var mailPool *email.Pool
var fromAddress string

// Init initializes the mailing
func Init(conf config.MailNotificationConf) {
	if !conf.Enabled {
		HTMLMailSender = noopSender{}
		PlainTextMailSender = noopSender{}
		return
	}
	fromAddress = conf.MailServer.FromAddress
	var err error
	mailPool, err = email.NewPool(
		fmt.Sprintf("%s:%d", conf.MailServer.Host, conf.MailServer.Port),
		4,
		smtp.PlainAuth("", conf.MailServer.Username, conf.MailServer.Password, conf.MailServer.Host),
	)
	if err != nil {
		log.WithError(err).Fatal("could not connect to email server")
	}
	log.Info("Connected to mail server")
	mailtemplates.Init()
}

var limits *multilimiter.MultiStore

func init() {
	var err error
	limits, err = multilimiter.NewDefaultMultiStore()
	if err != nil {
		panic(err)
	}
}

func sendEMail(mail *email.Email) error {
	err := errors.WithStack(mailPool.Send(mail, 10*time.Second))
	if err == nil {
		return nil
	}
	log.WithError(err).Error("error while sending mail")
	m := err.Error()
	if strings.Contains(m, "broken pipe") {
		// retry
		time.Sleep(300 * time.Millisecond)
		return errors.WithStack(mailPool.Send(mail, 10*time.Second))
	}
	return err
}

// SendEMail send the passed email.Email
func SendEMail(mail *email.Email) error {
	ok, reset, firstFailed, err := limits.Take(context.Background(), mail.To[0])
	if err != nil {
		return err
	}
	if ok {
		return sendEMail(mail)
	}
	if firstFailed {
		text := []byte(fmt.Sprintf(
			"You have reached your mail limit on this mytoken server. "+
				"The limit will reset in %s", time.Until(reset),
		))
		if mail.HTML != nil {
			mail.HTML = text
		} else {
			mail.Text = text
		}
		mail.Attachments = nil
		mail.Subject = "mail limit reached"
		err = sendEMail(mail)
		if err != nil {
			log.WithError(err).Error("error while sending mail limit mail")
		}
	}
	return errors.Errorf("mail limit reached; limit will reset in %s", time.Until(reset))
}

// Attachment is a type holding information about an email attachment
type Attachment struct {
	Reader      io.Reader
	Filename    string
	ContentType string
}
type attachmentMarshal struct {
	Data        []byte `json:"d"`
	Filename    string `json:"f"`
	ContentType string `json:"ct"`
}

func (a Attachment) MarshalJSON() ([]byte, error) {
	readerData, err := io.ReadAll(a.Reader)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	aa := attachmentMarshal{
		Data:        readerData,
		Filename:    a.Filename,
		ContentType: a.ContentType,
	}
	return json.Marshal(aa)
}

func (a *Attachment) UnmarshalJSON(data []byte) error {
	var aa attachmentMarshal
	if err := json.Unmarshal(data, &aa); err != nil {
		return errors.WithStack(err)
	}
	(*a).Filename = aa.Filename
	(*a).ContentType = aa.ContentType
	(*a).Reader = bytes.NewReader(aa.Data)
	return nil
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
func (noopSender) Send(_, _, _ string, _ ...Attachment) error {
	return nil
}

// SendTemplate implements the TemplateMailSender interface
func (noopSender) SendTemplate(_, _, _ string, _ any) error {
	return nil
}

// Send implements the MailSender interface
func (plainTextMailSender) Send(to, subject, text string, attachments ...Attachment) error {
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
func (htmlMailSender) Send(to, subject, text string, attachments ...Attachment) error {
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
func (icsMailSender) Send(to, subject, text string, attachments ...Attachment) error {
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
