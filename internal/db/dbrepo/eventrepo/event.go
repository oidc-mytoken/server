package eventrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

// EventDBObject holds information needed for storing an event in the database
type EventDBObject struct {
	api.Event
	Comment string
	MTID    mtid.MTID
	api.ClientMetaData
}

// Store stores the EventDBObject in the database
func (e *EventDBObject) Store(rlog log.Ext1FieldLogger, tx *sqlx.Tx) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(
				`CALL Event_Insert(?, ?, ?, ?, ?)`,
				e.MTID, e.Event, e.Comment, e.ClientMetaData.IP, e.ClientMetaData.UserAgent,
			)
			return errors.WithStack(err)
		},
	)
}
