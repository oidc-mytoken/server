package authcodeinforepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"

	"github.com/oidc-mytoken/api/v0"

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
	State *state.State
	pkg.AuthCodeFlowRequest
	PollingCode  bool
	CodeVerifier string
}

type authFlowInfo struct {
	State                   *state.State `db:"state_h"`
	pkg.AuthCodeFlowRequest `db:"request_json"`
	PollingCode             db.BitBool    `db:"polling_code"`
	CodeVerifier            db.NullString `db:"code_verifier"`
}

func (i *AuthFlowInfo) toAuthFlowInfo() *authFlowInfo {
	return &authFlowInfo{
		State:               i.State,
		AuthCodeFlowRequest: i.AuthCodeFlowRequest,
		PollingCode:         i.PollingCode != nil,
	}
}

func (i *authFlowInfo) toAuthFlowInfo() *AuthFlowInfoOut {
	return &AuthFlowInfoOut{
		State:               i.State,
		AuthCodeFlowRequest: i.AuthCodeFlowRequest,
		PollingCode:         bool(i.PollingCode),
		CodeVerifier:        i.CodeVerifier.String,
	}
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
			_, err := tx.Exec(
				`CALL AuthInfo_Insert(?, ?, ?, ?)`, store.State, store.AuthCodeFlowRequest,
				config.Get().Features.Polling.PollingCodeExpiresAfter, store.PollingCode,
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
			row := tx.QueryRowx(`CALL AuthInfo_Get(?)`, state)
			if err := row.Err(); err != nil {
				return errors.WithStack(err)
			}
			return errors.WithStack(
				row.Scan(
					&info.State, &info.AuthCodeFlowRequest, &info.PollingCode, &info.CodeVerifier,
				),
			)
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
			info, err := GetAuthFlowInfoByState(rlog, state)
			if err != nil {
				return err
			}
			info.Restrictions = r
			info.Capabilities = c
			info.SubtokenCapabilities = sc
			info.Rotation = rot
			info.Name = tokenName
			_, err = tx.Exec(
				`CALL AuthInfo_Update(?,?)`,
				state, info.AuthCodeFlowRequest,
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
