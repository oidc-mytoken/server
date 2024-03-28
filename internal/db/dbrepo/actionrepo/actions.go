package actionrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

// VerifyMail verifies a mail address
func VerifyMail(rlog log.Ext1FieldLogger, tx *sqlx.Tx, code string) (verified bool, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			result, err := tx.Exec(`CALL ActionCodes_VerifyMail(?)`, code)
			if err != nil {
				return errors.WithStack(err)
			}
			rows, err := result.RowsAffected()
			if err != nil {
				return errors.WithStack(err)
			}
			verified = rows == 1
			return DeleteCode(rlog, tx, code)
		},
	)
	return
}

// DeleteCode deletes a code
func DeleteCode(rlog log.Ext1FieldLogger, tx *sqlx.Tx, code string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL ActionCodes_Delete(?)`, code)
			return errors.WithStack(err)
		},
	)
}

// AddVerifyEmailCode adds a code for email verification to the database
func AddVerifyEmailCode(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, code string,
	expiresIn int,
) (err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err = tx.Exec(`CALL ActionCodes_AddVerifyMail(?,?,?)`, mtID, code, expiresIn)
			if err != nil {
				return errors.WithStack(err)
			}
			return err
		},
	)
	return
}

// AddRecreateTokenCode adds a code for token recreation to the database
func AddRecreateTokenCode(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, code string,
) (err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err = tx.Exec(`CALL ActionCodes_AddRecreateToken(?,?)`, mtID, code)
			if err != nil {
				return errors.WithStack(err)
			}
			return err
		},
	)
	return
}

// AddRemoveFromCalendarCode adds a code for removing a token from a calendar to the database
func AddRemoveFromCalendarCode(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, code, calendarName string,
) (err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err = tx.Exec(`CALL ActionCodes_AddRemoveFromCalendar(?,?,?)`, mtID, calendarName, code)
			if err != nil {
				return errors.WithStack(err)
			}
			return err
		},
	)
	return
}

// UseRemoveCalendarCode uses a calendar remove ActionCode to remove a token from a calendar and then deletes the
// code from the database
func UseRemoveCalendarCode(rlog log.Ext1FieldLogger, tx *sqlx.Tx, code string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL ActionCodes_UseRemoveFromCalendar(?)`, code)
			return errors.WithStack(err)
		},
	)
}
