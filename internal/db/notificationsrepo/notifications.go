package notificationsrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/endpoints/notification/pkg"
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

// NewNotification stores a new notification in the database
func NewNotification(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req pkg.SubscribeNotificationRequest,
	mtID mtid.MOMID, managementCode, ws string,
) error {
	if req.UserWide {
		return newUserWideNotification(rlog, tx, req, mtID, managementCode, ws)
	}
	return newMTNotification(rlog, tx, req, mtID, managementCode, ws)
}

func newUserWideNotification(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req pkg.SubscribeNotificationRequest,
	mtID mtid.MOMID, managementCode, ws string,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var nid uint64
			if err := errors.WithStack(
				tx.Get(
					&nid, `CALL Notifications_CreateUserWide(?,?,?,?)`, mtID, req.NotificationType,
					managementCode, db.NewNullString(ws),
				),
			); err != nil {
				return err
			}
			return linkNotificationClasses(rlog, tx, nid, req.NotificationClasses)
		},
	)
}

func newMTNotification(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req pkg.SubscribeNotificationRequest,
	mtID mtid.MOMID, managementCode, ws string,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var nid uint64
			if err := errors.WithStack(
				tx.Get(
					&nid, `CALL Notifications_CreateForMT(?,?,?,?)`, mtID, req.IncludeChildren, req.NotificationType,
					managementCode, ws,
				),
			); err != nil {
				return err
			}
			return linkNotificationClasses(rlog, tx, nid, req.NotificationClasses)
		},
	)
}

func linkNotificationClasses(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, nid uint64, classes []*api.NotificationClass,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			for _, nc := range classes {
				_, err := tx.Exec(`CALL Notifications_LinkClass(?,?)`, nid, nc.Name)
				if err != nil {
					return errors.WithStack(err)
				}
			}
			return nil
		},
	)
}
