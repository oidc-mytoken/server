package transfercoderepo

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/utils"
)

// TransferCode is a type used to transfer a token
type TransferCode struct {
	proxyToken
	Attributes transferCodeAttributes
}

type transferCodeAttributes struct {
	NewMT        db.BitBool
	ResponseType model.ResponseType
	MaxTokenLen  *int
}

// NewTransferCode creates a new TransferCode for the passed jwt
func NewTransferCode(jwt string, mID mtid.MTID, newMT bool, responseType model.ResponseType) (*TransferCode, error) {
	pt := newProxyToken(config.Get().Features.Polling.Len)
	if err := pt.SetJWT(jwt, mID); err != nil {
		return nil, err
	}
	transferCode := &TransferCode{
		proxyToken: *pt,
		Attributes: transferCodeAttributes{
			NewMT:        db.BitBool(newMT),
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
func CreatePollingCode(pollingCode string, responseType model.ResponseType, maxTokenLen int) *TransferCode {
	pt := createProxyToken(pollingCode)
	return &TransferCode{
		proxyToken: *pt,
		Attributes: transferCodeAttributes{
			NewMT:        true,
			ResponseType: responseType,
			MaxTokenLen:  utils.NewInt(maxTokenLen),
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
		_, err := tx.Exec(`CALL TransferCodeAttributes_Insert(?,?,?,?,?)`,
			tc.id, config.Get().Features.Polling.PollingCodeExpiresAfter, tc.Attributes.NewMT,
			tc.Attributes.ResponseType, tc.Attributes.MaxTokenLen)
		return errors.WithStack(err)
	})
}

// GetRevokeJWT returns a bool indicating if the linked jwt should also be revoked when this TransferCode is revoked or not
func (tc TransferCode) GetRevokeJWT(tx *sqlx.Tx) (bool, error) {
	var revokeMT db.BitBool
	err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		if err := tx.Get(&revokeMT, `CALL TransferCodeAttributes_GetRevokeJWT(?)`, tc.id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return errors.WithStack(err)
		}
		return nil
	})
	return bool(revokeMT), err
}
