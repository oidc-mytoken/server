package event

import (
	uuid "github.com/satori/go.uuid"
	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/model"
	pkg "github.com/zachmann/mytoken/internal/supertoken/event/pkg"
)

func LogEvent(event pkg.Event, stid uuid.UUID, metaData model.NetworkData) error {
	_, err := db.DB().Exec(`INSERT INTO ST_Events
(ST_id, event_id, comment, ip, user_agent)
VALUES(?, (SELECT id FROM Events WHERE event=?), ?, ?, ?)`, stid, event.String(), event.Comment, metaData.IP, metaData.UserAgent)
	return err
}
