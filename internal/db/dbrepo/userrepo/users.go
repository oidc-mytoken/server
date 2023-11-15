package userrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/mailing"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

type mailInfo struct {
	Mail           string `db:"email"`
	MailVerified   bool   `db:"email_verified"`
	PreferHTMLMail bool   `db:"prefer_html_mail"`
}

// GetMail returns the mail address and verification status for a user linked to a mytoken
func GetMail(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID) (data mailInfo, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&data, `CALL Users_GetMail(?)`, mtID))
		},
	)
	return
}

// GetTemplateMailSender returns a mailing.TemplateMailSender depending on the users preferred mime type
func GetTemplateMailSender(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID) (mailing.TemplateMailSender, error) {
	info, err := GetMail(rlog, tx, mtID)
	if err != nil {
		return nil, err
	}
	if info.PreferHTMLMail {
		return mailing.HTMLMailSender, nil
	}
	return mailing.PlainTextMailSender, nil
}

// ChangeEmail changes the user's email address
func ChangeEmail(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, newMail string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Users_ChangeMail(?,?)`, mtID, newMail)
			return errors.WithStack(err)
		},
	)
}

// ChangePreferredMailType changes the user's preferred email mimetype
func ChangePreferredMailType(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, preferHTML bool) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Users_ChangePreferredMailType(?,?)`, mtID, preferHTML)
			return errors.WithStack(err)
		},
	)
}
