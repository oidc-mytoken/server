package authcodeinforepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/utils"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
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
	Rotation             *api.Rotation
	ResponseType         model.ResponseType
	MaxTokenLen          int
	CodeVerifier         string
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
	Rotation             *api.Rotation
	ResponseType         model.ResponseType `db:"response_type"`
	MaxTokenLen          *int               `db:"max_token_len"`
	CodeVerifier         db.NullString      `db:"code_verifier"`
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
		Rotation:             i.Rotation,
		ResponseType:         i.ResponseType,
		MaxTokenLen:          utils.NewInt(i.MaxTokenLen),
	}
}

func (i *authFlowInfo) toAuthFlowInfo() *AuthFlowInfoOut {
	o := &AuthFlowInfoOut{
		State:                i.State,
		Issuer:               i.Issuer,
		Restrictions:         i.Restrictions,
		Capabilities:         i.Capabilities,
		SubtokenCapabilities: i.SubtokenCapabilities,
		Name:                 i.Name.String,
		PollingCode:          bool(i.PollingCode),
		Rotation:             i.Rotation,
		ResponseType:         i.ResponseType,
		CodeVerifier:         i.CodeVerifier.String,
	}
	if i.MaxTokenLen != nil {
		o.MaxTokenLen = *i.MaxTokenLen
	}
	return o
}

// Store stores the AuthFlowInfoIn in the database as well as the linked polling code if it exists
func (i *AuthFlowInfo) Store(rlog log.Ext1FieldLogger, tx *sqlx.Tx) error {
	rlog.Debug("Storing auth flow info")
	store := i.toAuthFlowInfo()
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if i.PollingCode != nil {
				if err := i.PollingCode.Store(rlog, tx); err != nil {
					return err
				}
			}
			_, err := tx.NamedExec(
				`CALL AuthInfo_Insert(:state_h, :iss, :restrictions, :capabilities, :subtoken_capabilities, :name, :expires_in, :polling_code, :rotation, :response_type, :max_token_len)`,
				store,
			)
			return errors.WithStack(err)
		},
	)
}

// GetAuthFlowInfoByState returns AuthFlowInfoIn by state
func GetAuthFlowInfoByState(rlog log.Ext1FieldLogger, state *state.State) (*AuthFlowInfoOut, error) {
	info := authFlowInfo{}
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&info, `CALL AuthInfo_Get(?)`, state))
		},
	); err != nil {
		return nil, err
	}
	return info.toAuthFlowInfo(), nil
}

// DeleteAuthFlowInfoByState deletes the AuthFlowInfoIn for a given state
func DeleteAuthFlowInfoByState(rlog log.Ext1FieldLogger, tx *sqlx.Tx, state *state.State) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL AuthInfo_Delete(?)`, state)
			return errors.WithStack(err)
		},
	)
}

// UpdateTokenInfoByState updates the stored AuthFlowInfo for the given state
func UpdateTokenInfoByState(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, state *state.State, r restrictions.Restrictions, c, sc api.Capabilities,
	rot *api.Rotation, tokenName string,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(
				`CALL AuthInfo_Update(?,?,?,?,?,?)`,
				state, r, c, sc, rot, tokenName,
			)
			return errors.WithStack(err)
		},
	)
}

// SetCodeVerifier stores the passed PKCE code verifier
func SetCodeVerifier(rlog log.Ext1FieldLogger, tx *sqlx.Tx, state *state.State, verifier string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL AuthInfo_SetCodeVerifier(?,?)`, state, verifier)
			return errors.WithStack(err)
		},
	)
}
