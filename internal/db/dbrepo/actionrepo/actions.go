package actionrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/unixtime"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
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
			return deleteCode(rlog, tx, code)
		},
	)
	return
}

// deleteCode deletes a code
func deleteCode(rlog log.Ext1FieldLogger, tx *sqlx.Tx, code string) error {
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

// AddScheduleNotificationCode adds a code for un-scheduling scheduled notifications
func AddScheduleNotificationCode(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, notificationID uint64, code string,
) (err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err = tx.Exec(`CALL ActionCodes_AddScheduleNotificationCode(?,?,?)`, code, notificationID, mtID)
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

// UseUnsubscribeFurtherNotificationsCode uses the ActionCode to unsubscribe from further scheduled notifications and
// deletes the code
func UseUnsubscribeFurtherNotificationsCode(rlog log.Ext1FieldLogger, tx *sqlx.Tx, code string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL ActionCodes_UnsubscribeFurtherScheduled(?)`, code)
			return errors.WithStack(err)
		},
	)
}

// RecreateData holds data stored in the database to enable re-creation of mytokens
type RecreateData struct {
	Name         db.NullString             `db:"name"`
	Issuer       string                    `db:"issuer"`
	Restrictions restrictions.Restrictions `db:"restrictions"`
	Capabilities api.Capabilities          `db:"capabilities"`
	Rotation     *api.Rotation             `db:"rotation"`
	Created      unixtime.UnixTime         `db:"created"`
}

// GetRecreateData returns the stored token recreation data linked to the passed code
func GetRecreateData(rlog log.Ext1FieldLogger, tx *sqlx.Tx, code string) (data RecreateData, found bool, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&data, `CALL ActionCodes_GetRecreateData(?)`, code))
		},
	)
	found, err = db.ParseError(err)
	return
}

// GetScheduledNotificationActionCode returns the action code for a scheduled notification
func GetScheduledNotificationActionCode(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, nid uint64) (
	code string,
	err error,
) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&code, `CALL ScheduledNotification_GetActionCode(?,?)`, nid, mtID))
		},
	)
	return
}
