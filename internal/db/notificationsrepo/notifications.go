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

// notificationInfoBaseWithClass is a type for holding information about a notification including class
type notificationInfoBaseWithClass struct {
	notificationInfoBase
	Class string `db:"class"`
}
type notificationInfoBase struct {
	api.NotificationInfoBase
	WebSocketPath db.NullString `db:"ws" json:"ws,omitempty"`
}

// GetNotificationsForMTAndClass checks for and returns the found notifications for a certain mytoken and
// notification class
func GetNotificationsForMTAndClass(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID,
	class api.NotificationClass,
) (notifications []api.NotificationInfoBase, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var dbNotifications []notificationInfoBase
			_, err = db.ParseError(
				tx.Select(
					&dbNotifications, `CALL Notifications_GetForMTAndClass(?,?)`, mtID,
					class.Name,
				),
			)
			if err != nil {
				return errors.WithStack(err)
			}
			for _, n := range dbNotifications {
				n.NotificationInfoBase.WebSocketPath = n.WebSocketPath.String
				notifications = append(notifications, n.NotificationInfoBase)
			}
			return nil
		},
	)
	return
}

// GetNotificationsForMT checks for and returns the found notifications for a certain mytoken
func GetNotificationsForMT(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID,
) (notifications []notificationInfoBaseWithClass, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err = db.ParseError(tx.Select(&notifications, `CALL Notifications_GetForMT(?)`, mtID))
			return errors.WithStack(err)
		},
	)
	return
}

// GetNotificationsForUser returns all found notifications for a user
func GetNotificationsForUser(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID,
) (notifications []api.NotificationInfo, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var withClass []notificationInfoBaseWithClass
			_, err = db.ParseError(tx.Select(&withClass, `CALL Notifications_GetForUser(?)`, mtID))
			if err != nil {
				return errors.WithStack(err)
			}
			notificationMap := make(map[uint64]api.NotificationInfo)
			for _, n := range withClass {
				nie, ok := notificationMap[n.NotificationID]
				if ok {
					nie.Classes = append(nie.Classes, api.NewNotificationClass(n.Class))
				} else {
					n.NotificationInfoBase.WebSocketPath = n.WebSocketPath.String
					nie = api.NotificationInfo{
						NotificationInfoBase: n.NotificationInfoBase,
						Classes:              []*api.NotificationClass{api.NewNotificationClass(n.Class)},
					}
					if !n.UserWide {
						if err = tx.Select(
							&nie.SubscribedTokens, `CALL Notifications_GetMTsForNotification(?)`,
							n.NotificationID,
						); err != nil {
							return err
						}
					}
				}
				notificationMap[nie.NotificationID] = nie
			}
			for _, n := range notificationMap {
				notifications = append(notifications, n)
			}
			return nil
		},
	)
	return
}

// GetNotificationForManagementCode returns the notification for a management code
func GetNotificationForManagementCode(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, managementCode string,
) (info *api.NotificationInfo, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var withClass []notificationInfoBaseWithClass
			found, err := db.ParseError(
				tx.Select(
					&withClass, `CALL Notifications_GetForManagementCode(?)`,
					managementCode,
				),
			)
			if err != nil {
				return errors.WithStack(err)
			}
			if !found || len(withClass) == 0 {
				info = nil
				return nil
			}
			// we can have multiple entries for the different classes, but they will all be for the same notification id
			info = &api.NotificationInfo{NotificationInfoBase: withClass[0].NotificationInfoBase}
			for _, n := range withClass {
				info.Classes = append(info.Classes, api.NewNotificationClass(n.Class))
			}
			if !info.UserWide {
				if err = tx.Select(
					&info.SubscribedTokens, `CALL Notifications_GetMTsForNotification(?)`,
					info.NotificationID,
				); err != nil {
					return err
				}
			}
			return nil
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
					&nid, `CALL Notifications_CreateForMT(?,?,?,?,?)`, mtID, req.IncludeChildren, req.NotificationType,
					managementCode, db.NewNullString(ws),
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
