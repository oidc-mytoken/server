package dbModels

import (
	"database/sql"
	"errors"

	"github.com/zachmann/mytoken/internal/model"

	"github.com/zachmann/mytoken/internal/db"
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
func CheckPollingCode(pollingCode string) (p PollingCodeStatus, err error) {
	if err = db.DB().Get(&p, `SELECT CURRENT_TIMESTAMP() <= expires_at AS valid FROM PollingCodes WHERE polling_code=?`, pollingCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil // polling code was not found, but this is fine
			return    // p.Found is false
		}
		return
	}
	return
}
