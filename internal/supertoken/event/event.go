package event

import (
	uuid "github.com/satori/go.uuid"
	"github.com/zachmann/mytoken/internal/db"
	pkg "github.com/zachmann/mytoken/internal/supertoken/event/pkg"
)

func LogEvent(event pkg.Event, stid uuid.UUID) error {
	//TODO
	ip := "192.168.0.31"
	userAgent := "go"

	_, err := db.DB().Exec(`INSERT INTO ST_Events
(ST_id, event_id, comment, ip, user_agent)
VALUES(?, (SELECT id FROM Events WHERE event=?), ?, ?, ?)`, stid, event.String(), event.Comment, ip, userAgent)
	return err
}
