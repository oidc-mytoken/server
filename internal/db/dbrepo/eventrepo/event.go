package eventrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
)

// EventDBObject holds information needed for storing an event in the database
type EventDBObject struct {
	*event.Event
	MTID mtid.MTID
	api.ClientMetaData
}

// Store stores the EventDBObject in the database
func (e *EventDBObject) Store(tx *sqlx.Tx) error {
	return db.RunWithinTransaction(
		tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(
				`CALL Event_Insert(?, ?, ?, ?, ?)`,
				e.MTID, e.Event.String(), e.Event.Comment, e.ClientMetaData.IP, e.ClientMetaData.UserAgent,
			)
			return errors.WithStack(err)
		},
	)
}
