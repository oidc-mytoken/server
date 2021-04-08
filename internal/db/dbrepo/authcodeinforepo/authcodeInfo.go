package authcodeinforepo

import (
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/pkg/api/v0"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
)

// AuthFlowInfo holds database information about a started authorization flow
type AuthFlowInfo struct {
	AuthFlowInfoOut
	PollingCode *transfercoderepo.TransferCode
}

// AuthFlowInfoOut holds database information about a started authorization flow
type AuthFlowInfoOut struct {
	State                *state.State
	Issuer               string
	Restrictions         restrictions.Restrictions
	Capabilities         api.Capabilities
	SubtokenCapabilities api.Capabilities
	Name                 string
	PollingCode          bool
}

type authFlowInfo struct {
	State                *state.State `db:"state_h"`
	Issuer               string       `db:"iss"`
	Restrictions         restrictions.Restrictions
	Capabilities         api.Capabilities
	SubtokenCapabilities api.Capabilities `db:"subtoken_capabilities"`
	Name                 db.NullString
	PollingCode          db.BitBool `db:"polling_code"`
	ExpiresIn            int64      `db:"expires_in"`
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
		PollingCode:          i.PollingCode != nil,
	}
}

func (i *authFlowInfo) toAuthFlowInfo() *AuthFlowInfoOut {
	return &AuthFlowInfoOut{
		State:                i.State,
		Issuer:               i.Issuer,
		Restrictions:         i.Restrictions,
		Capabilities:         i.Capabilities,
		SubtokenCapabilities: i.SubtokenCapabilities,
		Name:                 i.Name.String,
		PollingCode:          bool(i.PollingCode),
	}
}

// Store stores the AuthFlowInfoIn in the database as well as the linked polling code if it exists
func (i *AuthFlowInfo) Store(tx *sqlx.Tx) error {
	log.Debug("Storing auth flow info")
	store := i.toAuthFlowInfo()
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		if i.PollingCode != nil {
			if err := i.PollingCode.Store(tx); err != nil {
				return err
			}
		}
		_, err := tx.NamedExec(`INSERT INTO AuthInfo (state_h, iss, restrictions, capabilities, subtoken_capabilities, name, expires_in, polling_code) VALUES(:state_h, :iss, :restrictions, :capabilities, :subtoken_capabilities, :name, :expires_in, :polling_code)`, store)
		return err
	})
}

// GetAuthFlowInfoByState returns AuthFlowInfoIn by state
func GetAuthFlowInfoByState(state *state.State) (*AuthFlowInfoOut, error) {
	info := authFlowInfo{}
	if err := db.Transact(func(tx *sqlx.Tx) error {
		return tx.Get(&info, `SELECT state_h, iss, restrictions, capabilities, subtoken_capabilities, name, polling_code FROM AuthInfo WHERE state_h=? AND expires_at >= CURRENT_TIMESTAMP()`, state)
	}); err != nil {
		return nil, err
	}
	return info.toAuthFlowInfo(), nil
}

// DeleteAuthFlowInfoByState deletes the AuthFlowInfoIn for a given state
func DeleteAuthFlowInfoByState(tx *sqlx.Tx, state *state.State) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM AuthInfo WHERE state_h = ?`, state)
		return err
	})
}

func UpdateTokenInfoByState(tx *sqlx.Tx, state *state.State, r restrictions.Restrictions, c, sc api.Capabilities) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`UPDATE AuthInfo SET restrictions=?, capabilities=?, subtoken_capabilities=? WHERE state_h=?`, r, c, sc, state)
		return err
	})
}
