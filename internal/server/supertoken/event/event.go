package event

import (
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

	"github.com/zachmann/mytoken/internal/model"

	"github.com/zachmann/mytoken/internal/server/db/dbrepo/eventrepo"
	pkg "github.com/zachmann/mytoken/internal/server/supertoken/event/pkg"
)

// LogEvent logs an event to the database
func LogEvent(tx *sqlx.Tx, event *pkg.Event, stid uuid.UUID, clientMetaData model.ClientMetaData) error {
	return (&eventrepo.EventDBObject{
		Event:          event,
		STID:           stid,
		ClientMetaData: clientMetaData,
	}).Store(tx)
}
