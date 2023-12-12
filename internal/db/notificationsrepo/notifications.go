package notificationsrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

// ExpandNotificationsToChildrenIfApplicable checks if there is a notification subscription for the parent token that
// should include its children, and if so expands the subscription to also the just created child
func ExpandNotificationsToChildrenIfApplicable(rlog log.Ext1FieldLogger, tx *sqlx.Tx, parent, child mtid.MTID) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Notifications_ExpandToChildren(?,?)`, parent, child)
			return errors.WithStack(err)
		},
	)
}

// NotificationInfo is a type for holding information about a notification
type NotificationInfo struct {
	NotificationID uint64        `db:"id"`
	Type           string        `db:"type"`
	ManagementCode string        `db:"management_code"`
	WebSocketPath  db.NullString `db:"ws"`
	UserWide       bool          `db:"user_wide"`
}

// NotificationInfo is a type for holding information about a notification including class
type NotificationInfoWithClass struct {
	NotificationInfo
	Class string `db:"class"`
}

// GetNotificationsForMTAndClass checks for and returns the found notifications for a certain mytoken and
// notification class
func GetNotificationsForMTAndClass(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID,
	class api.NotificationClass,
) (notifications []NotificationInfo, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err = db.ParseError(
				tx.Select(
					&notifications, `CALL Notifications_GetForMTAndClass(?,?)`, mtID,
					class.Name,
				),
			)
			return errors.WithStack(err)
		},
	)
	return
}

// GetNotificationsForMT checks for and returns the found notifications for a certain mytoken
func GetNotificationsForMT(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID,
) (notifications []NotificationInfoWithClass, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err = db.ParseError(tx.Select(&notifications, `CALL Notifications_GetForMT(?)`, mtID))
			return errors.WithStack(err)
		},
	)
	return
}
