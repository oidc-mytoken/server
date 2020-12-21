package authcodeinforepo

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/supertoken/capabilities"
	"github.com/zachmann/mytoken/internal/supertoken/restrictions"
)

// AuthFlowInfo holds database information about a started authorization flow
type AuthFlowInfo struct {
	State                string
	Issuer               string
	Restrictions         restrictions.Restrictions
	Capabilities         capabilities.Capabilities
	SubtokenCapabilities capabilities.Capabilities
	Name                 string
	PollingCode          string
}

type authFlowInfo struct {
	State                string
	Issuer               string `db:"iss"`
	Restrictions         restrictions.Restrictions
	Capabilities         capabilities.Capabilities
	SubtokenCapabilities capabilities.Capabilities `db:"subtoken_capabilities"`
	Name                 sql.NullString
	PollingCodeID        *uint64        `db:"polling_code_id"`
	PollingCode          sql.NullString `db:"polling_code"`
	ExpiresIn            int64          `db:"expires_in"`
}

func (i *AuthFlowInfo) toAuthFlowInfo() *authFlowInfo {
	return &authFlowInfo{
		State:                i.State,
		Issuer:               i.Issuer,
		Restrictions:         i.Restrictions,
		Capabilities:         i.Capabilities,
		SubtokenCapabilities: i.SubtokenCapabilities,
		Name:                 db.NewNullString(i.Name),
		ExpiresIn:            config.Get().Features.Polling.PollingCodeExpiresAfter,
	}
}

func (i *authFlowInfo) toAuthFlowInfo() *AuthFlowInfo {
	return &AuthFlowInfo{
		State:                i.State,
		Issuer:               i.Issuer,
		Restrictions:         i.Restrictions,
		Capabilities:         i.Capabilities,
		SubtokenCapabilities: i.SubtokenCapabilities,
		Name:                 i.Name.String,
		PollingCode:          i.PollingCode.String,
	}
}

// Store stores the AuthFlowInfo in the database as well as the linked polling code if it exists
func (i *AuthFlowInfo) Store(tx *sqlx.Tx) error {
	log.Debug("Storing auth flow info")
	store := i.toAuthFlowInfo()
	storeFnc := func(tx *sqlx.Tx) error {
		if i.PollingCode != "" {
			res, err := tx.Exec(`INSERT INTO PollingCodes (polling_code, expires_in) VALUES(?, ?)`, i.PollingCode, config.Get().Features.Polling.PollingCodeExpiresAfter)
			if err != nil {
				return err
			}
			pid, err := res.LastInsertId()
			if err != nil {
				return err
			}
			upid := uint64(pid)
			store.PollingCodeID = &upid
		}
		_, err := tx.NamedExec(`INSERT INTO AuthInfo (state, iss, restrictions, capabilities, subtoken_capabilities, name, expires_in, polling_code_id) VALUES(:state, :iss, :restrictions, :capabilities, :subtoken_capabilities, :name, :expires_in, :polling_code_id)`, store)
		return err
	}
	return db.RunWithinTransaction(tx, storeFnc)
}

// GetAuthFlowInfoByState returns AuthFlowInfo by state
func GetAuthFlowInfoByState(state string) (*AuthFlowInfo, error) {
	info := authFlowInfo{}
	if err := db.DB().Get(&info, `SELECT state, iss, restrictions, capabilities, subtoken_capabilities, name, polling_code FROM AuthInfoV WHERE state=? AND expires_at >= CURRENT_TIMESTAMP()`, state); err != nil {
		return nil, err
	}
	return info.toAuthFlowInfo(), nil
}

// DeleteAuthFlowInfoByState deletes the AuthFlowInfo for a given state
func DeleteAuthFlowInfoByState(tx *sqlx.Tx, state string) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM AuthInfo WHERE state = ?`, state)
		return err
	})
}
