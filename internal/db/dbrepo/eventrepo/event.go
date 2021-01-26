package eventrepo

import (
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/server/db"
	event "github.com/zachmann/mytoken/internal/supertoken/event/pkg"
)

// EventDBObject holds information needed for storing an event in the database
type EventDBObject struct {
	*event.Event
	STID uuid.UUID
	model.ClientMetaData
}

// Store stores the EventDBObject in the database
func (e *EventDBObject) Store(tx *sqlx.Tx) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO ST_Events (ST_id, event_id, comment, ip, user_agent) VALUES(?, (SELECT id FROM Events WHERE event=?), ?, ?, ?)`,
			e.STID, e.Event.String(), e.Event.Comment, e.ClientMetaData.IP, e.ClientMetaData.UserAgent)
		return err
	})
}
