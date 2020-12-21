package transfercoderepo

import (
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/model"
	"github.com/zachmann/mytoken/internal/supertoken/event"
	pkg "github.com/zachmann/mytoken/internal/supertoken/event/pkg"
	"github.com/zachmann/mytoken/internal/utils"
)

// TransferCode holds information for transfer codes
type TransferCode struct {
	Code           string
	STID           uuid.UUID
	newST          bool
	clientMetaData model.ClientMetaData
}

// NewTransferCode creates a transfer code for the passed super token
func NewTransferCode(stid uuid.UUID, clientMetaData model.ClientMetaData, newST bool) *TransferCode {
	transferCode := utils.RandASCIIString(config.Get().Features.TransferCodes.Len)
	return &TransferCode{
		Code:           transferCode,
		STID:           stid,
		newST:          newST,
		clientMetaData: clientMetaData,
	}
}

// Store stores the TransferCode in the database
func (tc *TransferCode) Store(tx *sqlx.Tx) error {
	newST := 0
	if tc.newST {
		newST = 1
	}
	expiresIn := config.Get().Features.TransferCodes.ExpiresAfter
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		if _, err := tx.Exec(`INSERT INTO TransferCodes (transfer_code, ST_id, expires_in, new_st) VALUES(?,?,?,?)`, tc.Code, tc.STID, expiresIn, newST); err != nil {
			return err
		}
		if err := event.LogEvent(tx, &pkg.Event{
			Type: pkg.STEventTransferCodeCreated,
		}, tc.STID, tc.clientMetaData); err != nil {
			return err
		}
		return nil
	})
}
