package dbModels

import (
	"database/sql"
	"log"

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
	PollingCode  string `db:"polling_code"`
}

type authFlowInfo struct {
	State         string
	Issuer        string
	Restrictions  restrictions.Restrictions
	Capabilities  capabilities.Capabilities
	Name          sql.NullString
	PollingCodeID *uint64 `db:"polling_code_id"`
}

func newAuthFlowInfo(i *AuthFlowInfo) *authFlowInfo {
	return &authFlowInfo{
		State:        i.State,
		Issuer:       i.Issuer,
		Restrictions: i.Restrictions,
		Capabilities: i.Capabilities,
		Name:         db.NewNullString(i.Name),
	}
}

func (e *AuthFlowInfo) Store() error {
	log.Printf("Storing auth flow info")
	store := newAuthFlowInfo(e)
	return db.Transact(func(tx *sqlx.Tx) error {
		if e.PollingCode != "" {
			res, err := tx.Exec(`INSERT INTO PollingCodes (polling_code) VALUES(?)`, e.PollingCode)
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
		_, err := tx.NamedExec(`INSERT INTO AuthInfo (state, iss, restrictions, capabilities, name, polling_code_id) VALUES(:state, :issuer, :restrictions, :capabilities, :name, :polling_code_id)`, store)
		return err
	})
}

func GetAuthCodeInfoByState(state string) (info AuthFlowInfo, err error) {
	if e := db.DB().Get(&info, `SELECT * FROM AuthInfoV WHERE state=?`, state); e != nil {
		err = e
		return
	}
	return
}
