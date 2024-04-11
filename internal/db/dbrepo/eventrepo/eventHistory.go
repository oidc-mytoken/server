package eventrepo

import (
	"sort"

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
	MOMID          mtid.MOMID        `db:"MT_id" json:"mom_id"`
	Time           unixtime.UnixTime `db:"time" json:"time"`
}

// GetEventHistory returns the stored EventHistory for a mytoken
func GetEventHistory(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, incomingEvents EventHistory, ids ...any,
) (history EventHistory, err error) {
	history = incomingEvents
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			for _, id := range ids {
				var thisHistory EventHistory
				if err = errors.WithStack(tx.Select(&thisHistory.Events, `CALL EventHistory_Get(?)`, id)); err != nil {
					return err
				}
				history.Events = append(history.Events, thisHistory.Events...)
			}
			return nil
		},
	)
	sort.Slice(
		history.Events, func(i, j int) bool {
			return history.Events[i].Time > history.Events[j].Time
		},
	)
	return
}

// GetEventHistoryChildren returns the stored EventHistory for all children of a mytoken (
// not including the mytoken's own events)
func GetEventHistoryChildren(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, incomingEvents EventHistory, id any,
) (history EventHistory, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&history.Events, `CALL EventHistory_GetChildren(?)`, id))
		},
	)
	if len(incomingEvents.Events) > 0 {
		history.Events = append(incomingEvents.Events, history.Events...)
		sort.Slice(
			history.Events, func(i, j int) bool {
				return history.Events[i].Time > history.Events[j].Time
			},
		)
	}
	return
}

// GetPreviouslyUsedIPs returns a list of the ips that were previously used with a mytoken
func GetPreviouslyUsedIPs(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID) (ips []string, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err = db.ParseError(errors.WithStack(tx.Select(&ips, `CALL Events_getIPs(?)`, mtID)))
			return err
		},
	)
	return
}
