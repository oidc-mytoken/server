package dbModels

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/model"
)

// PollingCodeStatus holds information about the status of a polling code
type PollingCodeStatus struct {
	Found        bool
	Expired      bool
	ResponseType model.ResponseType
}

// Scan implements the Scanner interface
func (p *PollingCodeStatus) Scan(src interface{}) error {
	if src == nil {
		p.Found = false
		return nil
	}
	val := src.(int64)
	p.Found = true
	if val == 0 {
		p.Expired = true
	}
	return nil
}

// CheckPollingCode checks the passed polling code in the database
func CheckPollingCode(tx *sqlx.Tx, pollingCode string) (PollingCodeStatus, error) {
	var p PollingCodeStatus
	checkFnc := func(tx *sqlx.Tx) error {
		if err := tx.Get(&p, `SELECT CURRENT_TIMESTAMP() <= expires_at AS valid FROM PollingCodes WHERE polling_code=?`, pollingCode); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = nil  // polling code was not found, but this is fine
				return err // p.Found is false
			}
			return err
		}
		return nil
	}
	err := db.RunWithinTransaction(tx, checkFnc)
	return p, err
}
