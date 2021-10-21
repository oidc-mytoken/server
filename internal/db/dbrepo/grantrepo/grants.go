package grantrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/shared/model"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
)

// Enable enables a model.GrantType for the user of the passed mtid.MTID
func Enable(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myID mtid.MTID, grant model.GrantType) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Grants_Enable(?,?)`, myID, grant.String())
			return errors.WithStack(err)
		},
	)
}

// Disable disables a model.GrantType for the user of the passed mtid.MTID
func Disable(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myID mtid.MTID, grant model.GrantType) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Grants_Disable(?,?)`, myID, grant.String())
			return errors.WithStack(err)
		},
	)
}

// grantTypeInfo is a struct holding information indicating if a grant type is enabled or not
type grantTypeInfo struct {
	GrantType string     `db:"grant_type"`
	Enabled   db.BitBool `db:"enabled"`
}

func dbGrantInfoToAPIGrantInfo(dbGrants []grantTypeInfo) (apiGrants []api.GrantTypeInfo) {
	for _, g := range dbGrants {
		apiGrants = append(
			apiGrants, api.GrantTypeInfo{
				GrantType: g.GrantType,
				Enabled:   bool(g.Enabled),
			},
		)
	}
	return
}

// Get returns information about a user's enabled (and also some disabled) grant types
func Get(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myID mtid.MTID) ([]api.GrantTypeInfo, error) {
	var grants []grantTypeInfo
	if err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&grants, `CALL Grants_Get(?)`, myID))
		},
	); err != nil {
		return nil, err
	}
	return dbGrantInfoToAPIGrantInfo(grants), nil
}

// GrantEnabled checks if the passed model.GrantType is enabled for a user
func GrantEnabled(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myID mtid.MTID, grantType model.GrantType) (bool, error) {
	var enabled db.BitBool
	if err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&enabled, `CALL Grants_CheckEnabled(?,?)`, myID, grantType.String()))
		},
	); err != nil {
		return false, err
	}
	return bool(enabled), nil
}
