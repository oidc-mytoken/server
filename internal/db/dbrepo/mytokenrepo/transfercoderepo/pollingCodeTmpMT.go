package transfercoderepo

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

// TransferCodeStatus holds information about the status of a polling code
type TransferCodeStatus struct {
	Found             bool               `db:"found"`
	Expired           bool               `db:"expired"`
	ResponseType      model.ResponseType `db:"response_type"`
	ConsentDeclined   db.BitBool         `db:"consent_declined"`
	MaxTokenLen       *int               `db:"max_token_len"`
	SSHKeyFingerprint db.NullString      `db:"ssh_key_fp"`
}

// CheckTransferCode checks the passed polling code in the database
func CheckTransferCode(rlog log.Ext1FieldLogger, tx *sqlx.Tx, pollingCode string) (TransferCodeStatus, error) {
	pt := createProxyToken(pollingCode)
	var p TransferCodeStatus
	err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if err := tx.Get(&p, `CALL TransferCodes_GetStatus(?)`, pt.ID()); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					err = nil  // polling code was not found, but this is fine
					return err // p.Found is false
				}
				return errors.WithStack(err)
			}
			return nil
		},
	)
	return p, err
}

// PopTokenForTransferCode returns the decrypted token for a polling code and then deletes the entry
func PopTokenForTransferCode(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, pollingCode string, clientMetadata api.ClientMetaData,
) (
	jwt string, err error,
) {
	pt := createProxyToken(pollingCode)
	var valid bool
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			jwt, valid, err = pt.JWT(rlog, tx)
			if err != nil {
				return err
			}
			if !valid || jwt == "" {
				return nil
			}
			if err = pt.Delete(rlog, tx); err != nil {
				return err
			}
			return eventService.LogEvent(
				rlog, tx, eventService.MTEvent{
					Event: api.EventTransferCodeUsed,
					MTID:  pt.mtID,
				}, clientMetadata,
			)
		},
	)
	return
}

// LinkPollingCodeToMT links a pollingCode to a Mytoken
func LinkPollingCodeToMT(rlog log.Ext1FieldLogger, tx *sqlx.Tx, pollingCode, jwt string, mID mtid.MTID) error {
	pc := createProxyToken(pollingCode)
	if err := pc.SetJWT(jwt, mID); err != nil {
		return err
	}
	return pc.Update(rlog, tx)
}

// LinkPollingCodeToSSHKey links a pollingCode to an ssh public key
func LinkPollingCodeToSSHKey(rlog log.Ext1FieldLogger, tx *sqlx.Tx, pollingCode, sshKeyHash string) error {
	pc := createProxyToken(pollingCode)
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL TransferCodeAttributes_UpdateSSHKey(?,?)`, pc.ID(), sshKeyHash)
			return errors.WithStack(err)
		},
	)
}

// DeleteTransferCodeByState deletes a polling code
func DeleteTransferCodeByState(rlog log.Ext1FieldLogger, tx *sqlx.Tx, state *state.State) error {
	pc := createProxyToken(state.PollingCode(rlog))
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL ProxyTokens_Delete(?)`, pc.ID())
			return errors.WithStack(err)
		},
	)
}

// DeclineConsentByState updates the polling code attribute after the consent has been declined
func DeclineConsentByState(rlog log.Ext1FieldLogger, tx *sqlx.Tx, state *state.State) error {
	pc := createProxyToken(state.PollingCode(rlog))
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL TransferCodeAttributes_DeclineConsent(?)`, pc.ID())
			return errors.WithStack(err)
		},
	)
}
