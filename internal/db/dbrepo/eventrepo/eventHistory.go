package eventrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/utils/unixtime"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

// EventHistory is a slice of EventEntry
type EventHistory struct {
	api.EventHistory
	Events []EventEntry `json:"events"`
}

// EventEntry represents a mytoken event
type EventEntry struct {
	api.EventEntry `json:",inline"`
	MTID           mtid.MTID         `db:"MT_id" json:"-"`
	Time           unixtime.UnixTime `db:"time" json:"time"`
}

// GetEventHistory returns the stored EventHistory for a mytoken
func GetEventHistory(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id interface{}) (history EventHistory, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&history.Events, `CALL EventHistory_Get(?)`, id))
		},
	)
	return
}
