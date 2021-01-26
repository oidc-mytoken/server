package transfercoderepo

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/internal/server/db"
	"github.com/oidc-mytoken/server/pkg/model"
)

// TransferCodeStatus holds information about the status of a polling code
type TransferCodeStatus struct {
	Found        bool               `db:"found"`
	Expired      bool               `db:"expired"`
	ResponseType model.ResponseType `db:"response_type"`
	RedirectURL  sql.NullString     `db:"redirect"`
}

// CheckTransferCode checks the passed polling code in the database
func CheckTransferCode(tx *sqlx.Tx, pollingCode string) (TransferCodeStatus, error) {
	pt := createProxyToken(pollingCode)
	var p TransferCodeStatus
	err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		if err := tx.Get(&p, `SELECT 1 as found, CURRENT_TIMESTAMP() > expires_at AS expired, response_type, redirect FROM TransferCodes WHERE id=?`, pt.ID()); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = nil  // polling code was not found, but this is fine
				return err // p.Found is false
			}
			return err
		}
		return nil
	})
	return p, err
}

// PopTokenForTransferCode returns the decrypted token for a polling code and then deletes the entry
func PopTokenForTransferCode(tx *sqlx.Tx, pollingCode string) (jwt string, err error) {
	pt := createProxyToken(pollingCode)
	var valid bool
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		jwt, valid, err = pt.JWT(tx)
		if err != nil {
			return err
		}
		if !valid || len(jwt) == 0 {
			return nil
		}
		return pt.Delete(tx)
	})
	return
}

// LinkPollingCodeToST links a pollingCode to a SuperToken
func LinkPollingCodeToST(tx *sqlx.Tx, pollingCode, jwt string) error {
	pc := createProxyToken(pollingCode)
	if err := pc.SetJWT(jwt); err != nil {
		return err
	}
	return pc.Update(tx)
}

// DeleteTransferCodeByState deletes a polling code
func DeleteTransferCodeByState(tx *sqlx.Tx, state *state.State) error {
	pc := createProxyToken(state.PollingCode())
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM ProxyTokens WHERE id=?`, pc.ID())
		return err
	})
}
