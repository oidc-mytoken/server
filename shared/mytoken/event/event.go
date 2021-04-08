package event

import (
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/eventrepo"
	"github.com/oidc-mytoken/server/pkg/api/v0"
	pkg "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
)

type MTEvent struct {
	*pkg.Event
	MTID mtid.MTID
}

// LogEvent logs an event to the database
func LogEvent(tx *sqlx.Tx, event MTEvent, clientMetaData api.ClientMetaData) error {
	return (&eventrepo.EventDBObject{
		Event:          event.Event,
		MTID:           event.MTID,
		ClientMetaData: clientMetaData,
	}).Store(tx)
}

// LogEvents logs multiple events for the same token to the database
func LogEvents(tx *sqlx.Tx, events []MTEvent, clientMetaData api.ClientMetaData) error {
	for _, e := range events {
		if err := LogEvent(tx, e, clientMetaData); err != nil {
			return err
		}
	}
	return nil
}
