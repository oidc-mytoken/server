package transfercoderepo

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/shared/model"
	eventService "github.com/oidc-mytoken/server/shared/mytoken/event"
	event "github.com/oidc-mytoken/server/shared/mytoken/event/pkg"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
)

// TransferCodeStatus holds information about the status of a polling code
type TransferCodeStatus struct {
	Found           bool               `db:"found"`
	Expired         bool               `db:"expired"`
	ResponseType    model.ResponseType `db:"response_type"`
	ConsentDeclined db.BitBool         `db:"consent_declined"`
	MaxTokenLen     *int               `db:"max_token_len"`
}

// CheckTransferCode checks the passed polling code in the database
func CheckTransferCode(tx *sqlx.Tx, pollingCode string) (TransferCodeStatus, error) {
	pt := createProxyToken(pollingCode)
	var p TransferCodeStatus
	err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		if err := tx.Get(&p, `CALL TransferCodes_GetStatus(?)`, pt.ID()); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = nil  // polling code was not found, but this is fine
				return err // p.Found is false
			}
			return errors.WithStack(err)
		}
		return nil
	})
	return p, err
}

// PopTokenForTransferCode returns the decrypted token for a polling code and then deletes the entry
func PopTokenForTransferCode(tx *sqlx.Tx, pollingCode string, clientMetadata api.ClientMetaData) (jwt string, err error) {
	pt := createProxyToken(pollingCode)
	var valid bool
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		jwt, valid, err = pt.JWT(tx)
		if err != nil {
			return err
		}
		if !valid || jwt == "" {
			return nil
		}
		if err = pt.Delete(tx); err != nil {
			return err
		}
		return eventService.LogEvent(tx, eventService.MTEvent{
			Event: event.FromNumber(event.MTEventTransferCodeUsed, ""),
			MTID:  pt.mtID,
		}, clientMetadata)
	})
	return
}

// LinkPollingCodeToMT links a pollingCode to a Mytoken
func LinkPollingCodeToMT(tx *sqlx.Tx, pollingCode, jwt string, mID mtid.MTID) error {
	pc := createProxyToken(pollingCode)
	if err := pc.SetJWT(jwt, mID); err != nil {
		return err
	}
	return pc.Update(tx)
}

// DeleteTransferCodeByState deletes a polling code
func DeleteTransferCodeByState(tx *sqlx.Tx, state *state.State) error {
	pc := createProxyToken(state.PollingCode())
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`CALL ProxyTokens_Delete(?)`, pc.ID())
		return errors.WithStack(err)
	})
}

// DeclineConsentByState updates the polling code attribute after the consent has been declined
func DeclineConsentByState(tx *sqlx.Tx, state *state.State) error {
	pc := createProxyToken(state.PollingCode())
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`CALL TransferCodeAttributes_DeclineConsent(?)`, pc.ID())
		return errors.WithStack(err)
	})
}
