package dbModels

import (
	"database/sql"
	"log"

	"github.com/zachmann/mytoken/internal/config"

	"github.com/jmoiron/sqlx"

	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/supertoken/capabilities"
	"github.com/zachmann/mytoken/internal/supertoken/restrictions"
)

type AuthFlowInfo struct {
	State        string
	Issuer       string
	Restrictions restrictions.Restrictions
	Capabilities capabilities.Capabilities
	Name         string
	PollingCode  string
}

type authFlowInfo struct {
	State         string
	Issuer        string `db:"iss"`
	Restrictions  restrictions.Restrictions
	Capabilities  capabilities.Capabilities
	Name          sql.NullString
	PollingCodeID *uint64        `db:"polling_code_id"`
	PollingCode   sql.NullString `db:"polling_code"`
}

func (i *AuthFlowInfo) toAuthFlowInfo() *authFlowInfo {
	return &authFlowInfo{
		State:        i.State,
		Issuer:       i.Issuer,
		Restrictions: i.Restrictions,
		Capabilities: i.Capabilities,
		Name:         db.NewNullString(i.Name),
	}
}

func (i *authFlowInfo) toAuthFlowInfo() *AuthFlowInfo {
	return &AuthFlowInfo{
		State:        i.State,
		Issuer:       i.Issuer,
		Restrictions: i.Restrictions,
		Capabilities: i.Capabilities,
		Name:         i.Name.String,
		PollingCode:  i.PollingCode.String,
	}
}

func (i *AuthFlowInfo) Store() error {
	log.Printf("Storing auth flow info")
	store := i.toAuthFlowInfo()
	return db.Transact(func(tx *sqlx.Tx) error {
		if i.PollingCode != "" {
			res, err := tx.Exec(`INSERT INTO PollingCodes (polling_code, expires_in) VALUES(?, ?)`, i.PollingCode, config.Get().Polling.PollingCodeExpiresAfter)
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
		_, err := tx.NamedExec(`INSERT INTO AuthInfo (state, iss, restrictions, capabilities, name, polling_code_id) VALUES(:state, :iss, :restrictions, :capabilities, :name, :polling_code_id)`, store)
		return err
	})
}

func GetAuthCodeInfoByState(state string) (*AuthFlowInfo, error) {
	info := authFlowInfo{}
	if err := db.DB().Get(&info, `SELECT state, iss, restrictions, capabilities, name, polling_code FROM AuthInfoV WHERE state=?`, state); err != nil {
		return nil, err
	}
	return info.toAuthFlowInfo(), nil
}
