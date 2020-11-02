package dbModels

import (
	"database/sql"
	"log"

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
	State        string
	Issuer       string
	Restrictions restrictions.Restrictions
	Capabilities capabilities.Capabilities
	Name         sql.NullString
	PollingCode  sql.NullString `db:"polling_code"`
}

func newAuthFlowInfo(i *AuthFlowInfo) *authFlowInfo {
	return &authFlowInfo{
		State:        i.State,
		Issuer:       i.Issuer,
		Restrictions: i.Restrictions,
		Capabilities: i.Capabilities,
		Name:         db.NewNullString(i.Name),
		PollingCode:  db.NewNullString(i.PollingCode),
	}
}

func (e *AuthFlowInfo) Store() error {
	log.Printf("Storing auth flow info")
	store := newAuthFlowInfo(e)
	_, err := db.DB().NamedExec(`INSERT INTO AuthInfo (state, iss, restrictions, capabilities, name, polling_code) VALUES(:state, :issuer, :restrictions, :capabilities, :name, :polling_code)`, store)
	return err
}
