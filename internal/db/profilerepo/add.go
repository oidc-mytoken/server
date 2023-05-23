package profilerepo

import (
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
)

// AddProfile adds a profile for the passed group and name
func AddProfile(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, group, name string, payload json.RawMessage,
) error {
	id, err := uuid.NewV4()
	if err != nil {
		return errors.WithStack(err)
	}
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err = tx.Exec(`CALL Profiles_InsertProfiles(?,?,?,?)`, id, group, name, payload)
			return errors.WithStack(err)
		},
	)
}

// AddCapabilities adds a capability template for the passed group and name
func AddCapabilities(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, group, name string, payload json.RawMessage,
) error {
	id, err := uuid.NewV4()
	if err != nil {
		return errors.WithStack(err)
	}
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Profiles_InsertCapabilities(?,?,?,?)`, id, group, name, payload)
			return errors.WithStack(err)
		},
	)
}

// AddRestrictions adds a restrictions template for the passed group and name
func AddRestrictions(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, group, name string, payload json.RawMessage,
) error {
	id, err := uuid.NewV4()
	if err != nil {
		return errors.WithStack(err)
	}
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Profiles_InsertRestrictions(?,?,?,?)`, id, group, name, payload)
			return errors.WithStack(err)
		},
	)
}

// AddRotation adds a capability template for the passed group and name
func AddRotation(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, group, name string, payload json.RawMessage,
) error {
	id, err := uuid.NewV4()
	if err != nil {
		return errors.WithStack(err)
	}
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Profiles_InsertRotations(?,?,?,?)`, id, group, name, payload)
			return errors.WithStack(err)
		},
	)
}
