package notificationsrepo

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
)

func init() {
	rand.Seed(time.Now().UnixMilli())
}

// NotificationScheduler syncs with other mytoken server instances so that only one server does send notifications.
// Notifications are checked every minute if there are notification due and if one of the servers takes care of
// sending them.
func NotificationScheduler() {
	notMaxCounter := 1
	for {
		if !config.Get().Server.SingleServer && !determineMaster(&notMaxCounter) {
			continue
		}
		// do stuff as master
		fmt.Println("I now would send notifications if there are any")
		//TODO remove
		time.Sleep(time.Duration(rand.Float32()*120) * time.Second)
	}
}

func determineMaster(notMaxCounter *int) bool {
	t, _ := time.Parse(time.RFC822, time.Now().Format(time.RFC822))
	t1 := t.Add(time.Minute)
	t2 := t1.Add(30 * time.Second)
	time.Sleep(time.Until(t1))

	i := rand.Int() * *notMaxCounter
	if err := db.Transact(
		log.StandardLogger(), func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL NotificationSync_Write(?)`, i)
			return err
		},
	); err != nil {
		log.WithError(err).Error("notification_scheduler: determine_master: error writing to db")
		return false
	}
	time.Sleep(time.Until(t2))

	max := 0
	var syncNumbers []int
	if err := db.Transact(
		log.StandardLogger(), func(tx *sqlx.Tx) error {
			return tx.Select(&syncNumbers, `CALL NotificationSync_Read()`)
		},
	); err != nil {
		log.WithError(err).Error("notification_scheduler: determine_master: error reading from db")
		return false
	}
	for _, j := range syncNumbers {
		if j > max {
			max = j
		}
	}
	if max != i {
		*notMaxCounter++
		return false
	}
	*notMaxCounter = 1
	if err := db.Transact(
		log.StandardLogger(), func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL NotificationSync_Reset()`)
			return err
		},
	); err != nil {
		log.WithError(err).Error("notification_scheduler: determine_master: error resetting sync table")
	}
	return true
}
