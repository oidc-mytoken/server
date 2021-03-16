package eventrepo

import (
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/utils/unixtime"
)

type EventHistory []EventEntry

type EventEntry struct {
	MTID    mtid.MTID         `db:"MT_id" json:"-"`
	Event   string            `db:"event" json:"event"`
	Time    unixtime.UnixTime `db:"time" json:"time"`
	Comment string            `db:"comment" json:"comment,omitempty"`
	model.ClientMetaData
}

func GetEventHistory(tx *sqlx.Tx, id mtid.MTID) (history EventHistory, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Select(&history, `SELECT MT_id, event, time, comment, ip, user_agent FROM EventHistory WHERE MT_id=?`, id)
	})
	return
}
