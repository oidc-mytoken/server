package eventrepo

import (
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/pkg/api/v0"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/utils/unixtime"
)

type EventHistory []EventEntry

type EventEntry struct {
	api.EventEntry `json:",inline"`
	MTID           mtid.MTID         `db:"MT_id" json:"-"`
	Time           unixtime.UnixTime `db:"time" json:"time"`
}

func GetEventHistory(tx *sqlx.Tx, id mtid.MTID) (history EventHistory, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Select(&history, `SELECT MT_id, event, time, comment, ip, user_agent FROM EventHistory WHERE MT_id=?`, id)
	})
	return
}
