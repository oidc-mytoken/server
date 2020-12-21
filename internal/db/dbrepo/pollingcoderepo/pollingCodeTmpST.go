package pollingcoderepo

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

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

// LinkPollingCodeToST links a pollingCode to a SuperToken
func LinkPollingCodeToST(tx *sqlx.Tx, pollingCode string, stid uuid.UUID) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO TmpST (polling_code_id, ST_id) VALUES((SELECT id FROM PollingCodes WHERE polling_code = ?), ?)`, pollingCode, stid)
		return err
	})
}

// DeletePollingCode deletes a polling code
func DeletePollingCode(tx *sqlx.Tx, pollingCode string) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM PollingCodes WHERE polling_code=?`, pollingCode)
		return err
	})
}

// DeletePollingCodeByState deletes a polling code
func DeletePollingCodeByState(tx *sqlx.Tx, state string) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM PollingCodes WHERE id=(SELECT polling_code_id FROM AuthInfo WHERE state=?)`, state)
		return err
	})
}

// GetTokenForPollingCode returns the token for a pollingCode
func GetTokenForPollingCode(tx *sqlx.Tx, pollingCode string) (token string, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&token, `SELECT token FROM TmpST_by_polling_code WHERE polling_code=? AND CURRENT_TIMESTAMP() <= polling_code_expires_at`, pollingCode)
	})
	return
}

// PopTokenForPollingCode gets a token for the given polling code and then deletes entry from the db
func PopTokenForPollingCode(tx *sqlx.Tx, pollingCode string) (token string, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		token, err = GetTokenForPollingCode(tx, pollingCode)
		if err != nil {
			return err
		}
		return DeletePollingCode(tx, pollingCode)
	})
	return
}
