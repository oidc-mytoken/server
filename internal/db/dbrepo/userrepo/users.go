package userrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

// MailInfo holds information about a user's mail settings
type MailInfo struct {
	Mail           db.NullString `db:"email"`
	MailVerified   bool          `db:"email_verified"`
	PreferHTMLMail bool          `db:"prefer_html_mail"`
}

// GetMail returns the mail address and verification status for a user linked to a mytoken
func GetMail(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID) (data MailInfo, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&data, `CALL Users_GetMail(?)`, mtID))
		},
	)
	return
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

// SetEmail sets a user's email address
func SetEmail(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, mail string, mailVerified bool) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Users_SetMail(?,?,?)`, mtID, mail, mailVerified)
			return errors.WithStack(err)
		},
	)
}
