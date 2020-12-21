package event

import (
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

	"github.com/zachmann/mytoken/internal/model"
	pkg "github.com/zachmann/mytoken/internal/supertoken/event/pkg"
)

// LogEvent logs an event to the database
func LogEvent(tx *sqlx.Tx, event *pkg.Event, stid uuid.UUID, metaData model.ClientMetaData) error {
	_, err := tx.Exec(`INSERT INTO ST_Events
(ST_id, event_id, comment, ip, user_agent)
VALUES(?, (SELECT id FROM Events WHERE event=?), ?, ?, ?)`, stid, event.String(), event.Comment, metaData.IP, metaData.UserAgent)
	return err
}
