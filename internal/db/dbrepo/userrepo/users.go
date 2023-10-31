package userrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

// GetMail returns the mail address and verification status for a user linked to a mytoken
func GetMail(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID) (string, bool, error) {
	var data = struct {
		Mail         string `db:"email"`
		MailVerified bool   `db:"email_verified"`
	}{}
	err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&data, `CALL Users_GetMail(?)`, mtID))
		},
	)
	return data.Mail, data.MailVerified, err
}
