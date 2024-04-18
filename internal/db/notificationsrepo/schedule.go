package notificationsrepo

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/unixtime"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/actionrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/tree"
	"github.com/oidc-mytoken/server/internal/endpoints/actions/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

// Constants for scheduling classes
const (
	ScheduleClassExp = "exp"
)

// Constants for AdditionalInfo keys
const (
	AdditionalInfoKeyExpiresAt = "expires_at"
)

type jsonMap map[string]any

// Scan implements the sql.Scanner interface.
func (m *jsonMap) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	val := src.([]uint8)
	err := json.Unmarshal(val, m)
	return err
}

// Value implements the driver.Valuer interface
func (m jsonMap) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}
	return json.Marshal(m)
}

// ScheduledNotification holds information about a notification that is scheduled
type ScheduledNotification struct {
	ScheduleID     uint64    `db:"id"`
	DueTime        time.Time `db:"due_time"`
	MTID           mtid.MTID `db:"MT_id"`
	Class          string    `db:"class"`
	AdditionalInfo jsonMap   `db:"additional_info"`
	NotificationID uint64    `db:"notification_id"`
	NotificationInfoBase
}

// PopOneScheduledNotification pops one ScheduledNotification from the database that is due
func PopOneScheduledNotification(rlog log.Ext1FieldLogger, tx *sqlx.Tx) (*ScheduledNotification, error) {
	n := &ScheduledNotification{}
	err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			err := errors.WithStack(tx.Get(n, `CALL PopOneDueScheduledNotification()`))
			if err != nil {
				_, err = db.ParseError(err)
				if err == nil {
					n = nil
				}
			}
			return err
		},
	)
	return n, err
}

var notificationIntervalsBeforeExpiration = []uint64{
	30 * 24 * 60 * 60,
	14 * 24 * 60 * 60,
	7 * 24 * 60 * 60,
	3 * 24 * 60 * 60,
	1 * 24 * 60 * 60,
	3 * 60 * 60,
	0,
}

// ScheduleExpirationNotificationsIfNeeded schedules all the needed expiration notifications if an exp notification
// is enabled for this mytoken
func ScheduleExpirationNotificationsIfNeeded(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, expiresAt unixtime.UnixTime, createdAt unixtime.UnixTime,
) error {
	notifications, err := GetNotificationsForMTAndClass(rlog, tx, mtID, api.NotificationClassExpiration)
	if err != nil {
		return err
	}
	if len(notifications) == 0 {
		return nil
	}
	for _, n := range notifications {
		if err = ScheduleExpirationNotifications(rlog, tx, n.NotificationID, mtID, expiresAt, createdAt); err != nil {
			return err
		}
	}
	return nil
}

// ScheduleExpirationNotifications schedules all the needed expiration notifications for this mytoken
func ScheduleExpirationNotifications(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, nid uint64, mtID mtid.MTID,
	expiresAt unixtime.UnixTime, createdAt unixtime.UnixTime,
) error {
	now := unixtime.Now()
	if expiresAt == 0 || expiresAt < now {
		return nil
	}
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {

			tokenLifetime := uint64(expiresAt - createdAt)
			halfTokenLifetime := tokenLifetime / 2
			var numberScheduledNotifications uint
			for _, secondsBeforeExpiration := range notificationIntervalsBeforeExpiration {
				if secondsBeforeExpiration > halfTokenLifetime {
					continue
				}
				notifyAt := expiresAt - unixtime.UnixTime(secondsBeforeExpiration)
				if notifyAt < now {
					continue
				}
				data := ScheduledNotification{
					DueTime:        notifyAt.Time(),
					MTID:           mtID,
					Class:          ScheduleClassExp,
					AdditionalInfo: jsonMap{AdditionalInfoKeyExpiresAt: expiresAt},
					NotificationID: nid,
				}
				_, err := tx.NamedExec(
					`CALL NotificationScheduleAdd(:due_time,:notification_id,:MT_id,:class,:additional_info)`, data,
				)
				if err = errors.WithStack(err); err != nil {
					return err
				}
				numberScheduledNotifications++
			}
			if numberScheduledNotifications > 0 {
				_, err := actionrepo.GetScheduledNotificationActionCode(rlog, tx, mtID, nid)
				var found bool
				found, err = db.ParseError(err)
				if err != nil {
					return err
				}
				if !found {
					if err = actionrepo.AddScheduleNotificationCode(rlog, tx, mtID, nid, pkg.NewCode()); err != nil {
						return err
					}
				}
			}
			return nil
		},
	)
}

// DeleteScheduledExpirationNotifications deletes all scheduled notifications for a Notification
func DeleteScheduledExpirationNotifications(rlog log.Ext1FieldLogger, tx *sqlx.Tx, nid uint64) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL NotificationSchedule_DeleteExpirations(?)`, nid)
			return errors.WithStack(err)
		},
	)
}

// DeleteScheduledExpirationNotificationsForMT deletes all scheduled notifications for a Notification and a certain
// mytoken
func DeleteScheduledExpirationNotificationsForMT(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, nid uint64, mtID mtid.MTID,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL NotificationSchedule_DeleteExpirationsForMT(?,?)`, nid, mtID)
			return errors.WithStack(err)
		},
	)
}

// AddScheduledExpirationNotifications adds all the needed expiration notifications for all the mytokens linked to a
// Notification
func AddScheduledExpirationNotifications(rlog log.Ext1FieldLogger, tx *sqlx.Tx, info NotificationInfoBase) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if info.UserWide {
				tokens, err := tree.AllTokensByUID(rlog, tx, info.UID)
				if err != nil {
					return err
				}
				for _, mt := range tokens {
					if err = ScheduleExpirationNotifications(
						rlog, tx, info.NotificationID, mt.ID, mt.ExpiresAt, mt.CreatedAt,
					); err != nil {
						return err
					}
				}
			} else {
				var mtids []mtid.MTID
				if err := tx.Select(
					&mtids, `CALL Notifications_GetMTsForNotification(?)`, info.NotificationID,
				); err != nil {
					return err
				}
				for _, mtID := range mtids {
					tokenEntry, err := tree.SingleTokenEntry(rlog, tx, mtID)
					if err != nil {
						return err
					}
					if err = ScheduleExpirationNotifications(
						rlog, tx, info.NotificationID, mtID, tokenEntry.ExpiresAt, tokenEntry.CreatedAt,
					); err != nil {
						return err
					}
				}
			}
			return nil
		},
	)
}
