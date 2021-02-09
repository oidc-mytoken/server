package transfercoderepo

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/pkg/model"
)

// TransferCode is a type used to transfer a token
type TransferCode struct {
	proxyToken
	Attributes transferCodeAttributes
}

type transferCodeAttributes struct {
	NewST db.BitBool
	model.ResponseType
}

// NewTransferCode creates a new TransferCode for the passed jwt
func NewTransferCode(jwt string, newST bool, responseType model.ResponseType) (*TransferCode, error) {
	pt := newProxyToken(config.Get().Features.Polling.Len)
	if err := pt.SetJWT(jwt); err != nil {
		return nil, err
	}
	transferCode := &TransferCode{
		proxyToken: *pt,
		Attributes: transferCodeAttributes{
			NewST:        db.BitBool(newST),
			ResponseType: responseType,
		},
	}
	return transferCode, nil
}

// ParseTransferCode creates a new transfer code from a transfer code string
func ParseTransferCode(token string) *TransferCode {
	return &TransferCode{proxyToken: *parseProxyToken(token)}
}

// CreatePollingCode creates a polling code
func CreatePollingCode(pollingCode string, responseType model.ResponseType) *TransferCode {
	pt := createProxyToken(pollingCode)
	return &TransferCode{
		proxyToken: *pt,
		Attributes: transferCodeAttributes{
			NewST:        true,
			ResponseType: responseType,
		},
	}
}

// Store stores the TransferCode in the database
func (tc TransferCode) Store(tx *sqlx.Tx) error {
	log.Debug("Storing transfer code")
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		if err := tc.proxyToken.Store(tx); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO TransferCodesAttributes (id, expires_in,  revoke_ST, response_type) VALUES(?,?,?,?)`, tc.id, config.Get().Features.Polling.PollingCodeExpiresAfter, tc.Attributes.NewST, tc.Attributes.ResponseType)
		return err
	})
}

// GetRevokeJWT returns a bool indicating if the linked jwt should also be revoked when this TransferCode is revoked or not
func (tc TransferCode) GetRevokeJWT(tx *sqlx.Tx) (bool, error) {
	var revokeST db.BitBool
	err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		if err := tx.Get(&revokeST, `SELECT revoke_ST FROM TransferCodesAttributes WHERE id=?`, tc.id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}
		return nil
	})
	return bool(revokeST), err
}
