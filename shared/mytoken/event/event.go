package event

import (
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/eventrepo"
	pkg "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
)

// MTEvent is type for mytoken events
type MTEvent struct {
	*pkg.Event
	MTID mtid.MTID
}

// LogEvent logs an event to the database
func LogEvent(rlog log.Ext1FieldLogger, tx *sqlx.Tx, event MTEvent, clientMetaData api.ClientMetaData) error {
	return (&eventrepo.EventDBObject{
		Event:          event.Event,
		MTID:           event.MTID,
		ClientMetaData: clientMetaData,
	}).Store(rlog, tx)
}

// LogEvents logs multiple events for the same token to the database
func LogEvents(rlog log.Ext1FieldLogger, tx *sqlx.Tx, events []MTEvent, clientMetaData api.ClientMetaData) error {
	for _, e := range events {
		if err := LogEvent(rlog, tx, e, clientMetaData); err != nil {
			return err
		}
	}
	return nil
}
