package event

import (
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/eventrepo"
	"github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	notifier "github.com/oidc-mytoken/server/internal/notifier/client"
)

// LogEvent logs an event to the database
func LogEvent(rlog log.Ext1FieldLogger, tx *sqlx.Tx, event pkg.MTEvent) error {
	if err := (&eventrepo.EventDBObject{
		Event:          event.Event,
		Comment:        event.Comment,
		MTID:           event.MTID,
		ClientMetaData: event.ClientMetaData,
	}).Store(rlog, tx); err != nil {
		return err
	}
	return notifier.SendNotificationsForEvent(rlog, tx, event)
}

// LogEvents logs multiple events for the same token to the database
func LogEvents(rlog log.Ext1FieldLogger, tx *sqlx.Tx, events []pkg.MTEvent) error {
	for _, e := range events {
		if err := LogEvent(rlog, tx, e); err != nil {
			return err
		}
	}
	return nil
}
