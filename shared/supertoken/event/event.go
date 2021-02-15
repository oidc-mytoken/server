package event

import (
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/eventrepo"
	"github.com/oidc-mytoken/server/internal/model"
	pkg "github.com/oidc-mytoken/server/shared/supertoken/event/pkg"
	"github.com/oidc-mytoken/server/shared/supertoken/pkg/stid"
)

// LogEvent logs an event to the database
func LogEvent(tx *sqlx.Tx, event *pkg.Event, stid stid.STID, clientMetaData model.ClientMetaData) error {
	return (&eventrepo.EventDBObject{
		Event:          event,
		STID:           stid,
		ClientMetaData: clientMetaData,
	}).Store(tx)
}
