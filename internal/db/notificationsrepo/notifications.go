package notificationsrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/notificationsrepo/calendarrepo"
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

// NotificationInfoBaseWithClass is a type for holding information about a notification including class
type NotificationInfoBaseWithClass struct {
	NotificationInfoBase
	Class string `db:"class"`
}

// NotificationInfoBase is a type for holding information about a notification
type NotificationInfoBase struct {
	api.NotificationInfoBase
	WebSocketPath db.NullString `db:"ws" json:"ws,omitempty"`
	UID           uint64        `db:"uid" json:"-"`
}

// ManagementCodeNotificationInfoResponse extens api.ManagementCodeNotificationInfoResponse with an uid (not for json)
type ManagementCodeNotificationInfoResponse struct {
	api.ManagementCodeNotificationInfoResponse
	UID uint64 `db:"uid" json:"-"`
}

// oidcInfo holds OIDC issuer and subject for a user
type oidcInfo struct {
	Iss string `db:"iss"`
	Sub string `db:"sub"`
}

// GetNotificationsForMTAndClass checks for and returns the found notifications for a certain mytoken and
// notification class
func GetNotificationsForMTAndClass(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID,
	class *api.NotificationClass,
) (notifications []api.NotificationInfoBase, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var dbNotifications []NotificationInfoBase
			_, err = db.ParseError(
				tx.Select(
					&dbNotifications, `CALL Notifications_GetForMTAndClass(?,?)`, mtID, class.Name,
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
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID any,
) (notifications []NotificationInfoBaseWithClass, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err = db.ParseError(tx.Select(&notifications, `CALL Notifications_GetForMT(?)`, mtID))
			return errors.WithStack(err)
		},
	)
	return
}

// GetNotificationsAndCalendarsForMT obtains all api.NotificationInfo and api.
// CalendarInfo for a mytoken from the database
func GetNotificationsAndCalendarsForMT(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID any,
) (notifications []api.NotificationInfo, calendars []api.CalendarInfo, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			calendars, err = calendarrepo.ListCalendarsForMT(rlog, tx, mtID)
			if err != nil {
				return err
			}
			ns, err := GetNotificationsForMT(rlog, tx, mtID)
			if err != nil {
				return err
			}
			notifications, err = notificationInfoBaseWithClassToNotificationInfo(rlog, tx, ns)
			return err
		},
	)
	return
}

func notificationInfoBaseWithClassToNotificationInfo(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx,
	in []NotificationInfoBaseWithClass,
) (
	out []api.NotificationInfo,
	err error,
) {
	notificationMap := make(map[uint64]api.NotificationInfo)
	var ids []uint64
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			for _, n := range in {
				nie, ok := notificationMap[n.NotificationID]
				if ok {
					nie.Classes = append(nie.Classes, api.NewNotificationClass(n.Class))
				} else {
					ids = append(ids, n.NotificationID)
					n.NotificationInfoBase.NotificationInfoBase.WebSocketPath = n.WebSocketPath.String
					nie = api.NotificationInfo{
						NotificationInfoBase: n.NotificationInfoBase.NotificationInfoBase,
						Classes:              api.NotificationClasses{api.NewNotificationClass(n.Class)},
					}
					if !n.UserWide {
						// Only return directly subscribed tokens (not tag-based)
						if err = tx.Select(
							&nie.SubscribedTokens, `CALL Notifications_GetDirectMTsForNotification(?)`,
							n.NotificationID,
						); err != nil {
							return err
						}
					}
					// Fetch tags for this notification
					if nie.Tags, err = GetNotificationTags(rlog, tx, n.NotificationID); err != nil {
						return err
					}
				}
				notificationMap[nie.NotificationID] = nie
			}
			return nil
		},
	)
	if err != nil {
		return
	}
	for _, id := range ids {
		out = append(out, notificationMap[id])
	}
	return
}

// GetNotificationsForUser returns all found notifications for a user
func GetNotificationsForUser(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID,
) (notifications []api.NotificationInfo, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var withClass []NotificationInfoBaseWithClass
			_, err = db.ParseError(tx.Select(&withClass, `CALL Notifications_GetForUser(?)`, mtID))
			if err != nil {
				return errors.WithStack(err)
			}
			notifications, err = notificationInfoBaseWithClassToNotificationInfo(rlog, tx, withClass)
			return err
		},
	)
	return
}

// GetNotificationForManagementCode returns the notification for a management code
func GetNotificationForManagementCode(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, managementCode string,
) (info *ManagementCodeNotificationInfoResponse, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var withClass []NotificationInfoBaseWithClass
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
			info = &ManagementCodeNotificationInfoResponse{
				ManagementCodeNotificationInfoResponse: api.ManagementCodeNotificationInfoResponse{
					NotificationInfo: api.NotificationInfo{
						NotificationInfoBase: withClass[0].NotificationInfoBase.NotificationInfoBase,
					},
				},
				UID: withClass[0].UID,
			}
			for _, n := range withClass {
				info.Classes = append(info.Classes, api.NewNotificationClass(n.Class))
			}
			if !info.UserWide {
				// Only return directly subscribed tokens (not tag-based)
				if err = errors.WithStack(
					tx.Select(
						&info.SubscribedTokens, `CALL Notifications_GetDirectMTsForNotification(?)`,
						info.NotificationID,
					),
				); err != nil {
					if _, err = db.ParseError(err); err != nil {
						return err
					}
				}
			}
			// Fetch tags for this notification
			if info.Tags, err = GetNotificationTags(rlog, tx, info.NotificationID); err != nil {
				return err
			}
			// Fetch OIDC issuer and subject
			var oi oidcInfo
			if err := errors.WithStack(
				tx.Get(
					&oi, `CALL GetOIDCInfoForManagementCode(?)`, managementCode,
				),
			); err != nil {
				return err
			}
			info.OIDCIssuer = oi.Iss
			info.OIDCSub = oi.Sub
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
	// If tags are provided and no explicit mom_id, create notification without linking a token
	// The token will be associated via tags instead
	if len(req.Tags) > 0 && !req.MomID.HashValid() {
		return newTagOnlyNotification(rlog, tx, req, mtID, managementCode, ws)
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

func newTagOnlyNotification(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, req pkg.SubscribeNotificationRequest,
	mtID mtid.MOMID, managementCode, ws string,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var nid uint64
			if err := errors.WithStack(
				tx.Get(
					&nid, `CALL Notifications_CreateWithoutMT(?,?,?,?)`, mtID, req.NotificationType,
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
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, nid uint64, classes api.NotificationClasses,
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

// AddTokenToNotification subscribes an additional token (and possibly its children) to a notification
func AddTokenToNotification(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, notificationID uint64, mtID mtid.MOMID, includeChildren bool,
) (err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if includeChildren {
				_, err = tx.Exec(`CALL Notifications_LinkMTWithChildren(?,?)`, mtID, notificationID)
			} else {
				_, err = tx.Exec(`CALL Notifications_LinkMT(?,?,?)`, mtID, notificationID, 0)
			}
			return errors.WithStack(err)

		},
	)
	return
}

// RemoveTokenFromNotification unsubscribes a token (and possibly its children) from a notification
func RemoveTokenFromNotification(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, notificationID uint64, mtID mtid.MOMID,
) (err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err = tx.Exec(`CALL Notifications_UnlinkMT(?,?)`, mtID, notificationID)
			return errors.WithStack(err)
		},
	)
	return
}

// UpdateNotificationClasses updates the notification classes for a notification
func UpdateNotificationClasses(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, notificationID uint64, newClasses api.NotificationClasses,
) (err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err = tx.Exec(`CALL Notifications_ClearNotificationClasses(?)`, notificationID)
			if err != nil {
				return errors.WithStack(err)
			}
			return linkNotificationClasses(rlog, tx, notificationID, newClasses)
		},
	)
	return
}

// Delete deletes the notification for a managementCode
func Delete(rlog log.Ext1FieldLogger, tx *sqlx.Tx, managementCode string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Notifications_DeleteByManagementCode(?)`, managementCode)
			return errors.WithStack(err)
		},
	)
}

// MytokenSubscribeOrCreateNotificationWithClasses checks if a notification
// already exists with the request type and api.
// NotificationClasses; if yes the mytoken subscribes to this notification,
// if not a new notification is created.
func MytokenSubscribeOrCreateNotificationWithClasses(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx,
	req api.CreateMytokenSubscribeNotificationInfos,
	mtID mtid.MTID,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			notifications, err := GetNotificationsForUser(rlog, tx, mtID)
			if err != nil {
				return err
			}
			for _, n := range notifications {
				if n.UserWide {
					continue
				}
				if n.Type != req.NotificationType {
					continue
				}
				if n.Classes.Equals(req.NotificationClasses) {
					return AddTokenToNotification(
						rlog, tx, n.NotificationID, mtID.MomID(), req.IncludeChildren,
					)
				}
			}
			// No existing notification found with the requested classes,
			// create a new one
			var nid uint64
			if err = errors.WithStack(
				tx.Get(
					&nid, `CALL Notifications_CreateForMT(?,?,?,?,?)`, mtID, req.IncludeChildren, req.NotificationType,
					utils.RandASCIIString(64), db.NewNullString(""),
				),
			); err != nil {
				return err
			}
			return linkNotificationClasses(rlog, tx, nid, req.NotificationClasses)
		},
	)
}

// LinkTags clears and links the provided tags to a notification
func LinkTags(rlog log.Ext1FieldLogger, tx *sqlx.Tx, notificationID uint64, tags []api.Tag) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Notifications_ClearTags(?)`, notificationID)
			if err != nil {
				return errors.WithStack(err)
			}
			for _, tag := range tags {
				_, err = tx.Exec(`CALL Notifications_LinkTag(?,?)`, notificationID, tag)
				if err != nil {
					return errors.WithStack(err)
				}
			}
			return nil
		},
	)
}

// AddTag adds a tag to a notification
func AddTag(rlog log.Ext1FieldLogger, tx *sqlx.Tx, notificationID uint64, tag string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Notifications_LinkTag(?,?)`, notificationID, tag)
			return errors.WithStack(err)
		},
	)
}

// RemoveTag removes a tag from a notification
func RemoveTag(rlog log.Ext1FieldLogger, tx *sqlx.Tx, notificationID uint64, tag string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Notifications_UnlinkTag(?,?)`, notificationID, tag)
			return errors.WithStack(err)
		},
	)
}

// GetNotificationTags returns the api.TagInfos for a notification
func GetNotificationTags(rlog log.Ext1FieldLogger, tx *sqlx.Tx, notificationID uint64) (tags []api.TagInfo, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&tags, `CALL Notifications_GetTags(?)`, notificationID))
		},
	)
	return
}
