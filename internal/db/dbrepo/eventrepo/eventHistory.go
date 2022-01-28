package eventrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/utils/unixtime"
)

// EventHistory is type for multiple EventEntry
type EventHistory []EventEntry

// EventEntry represents a mytoken event
type EventEntry struct {
	api.EventEntry `json:",inline"`
	MTID           mtid.MTID         `db:"MT_id" json:"-"`
	Time           unixtime.UnixTime `db:"time" json:"time"`
}

// GetEventHistory returns the stored EventHistory for a mytoken
func GetEventHistory(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID) (history EventHistory, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&history, `CALL EventHistory_Get(?)`, id))
		},
	)
	return
}
