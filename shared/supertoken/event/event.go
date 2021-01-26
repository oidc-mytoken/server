package event

import (
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/eventrepo"
	"github.com/oidc-mytoken/server/internal/model"
	pkg "github.com/oidc-mytoken/server/shared/supertoken/event/pkg"
)

// LogEvent logs an event to the database
func LogEvent(tx *sqlx.Tx, event *pkg.Event, stid uuid.UUID, clientMetaData model.ClientMetaData) error {
	return (&eventrepo.EventDBObject{
		Event:          event,
		STID:           stid,
		ClientMetaData: clientMetaData,
	}).Store(tx)
}
