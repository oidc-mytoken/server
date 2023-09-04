package profilerepo

import (
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
)

// UpdateProfile updates a profile for the passed group and id
func UpdateProfile(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string, id uuid.UUID, name string, payload json.RawMessage,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Profiles_UpdateProfiles(?,?,?,?)`, id, group, name, payload)
			return errors.WithStack(err)
		},
	)
}

// UpdateCapabilities updates a capability template for the passed group and id
func UpdateCapabilities(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string, id uuid.UUID, name string, payload json.RawMessage,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Profiles_UpdateCapabilities(?,?,?,?)`, id, group, name, payload)
			return errors.WithStack(err)
		},
	)
}

// UpdateRestrictions updates a restrictions template for the passed group and id
func UpdateRestrictions(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string, id uuid.UUID, name string, payload json.RawMessage,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Profiles_UpdateRestrictions(?,?,?,?)`, id, group, name, payload)
			return errors.WithStack(err)
		},
	)
}

// UpdateRotation updates a capability template for the passed group and id
func UpdateRotation(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string, id uuid.UUID, name string, payload json.RawMessage,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Profiles_UpdateRotations(?,?,?,?)`, id, group, name, payload)
			return errors.WithStack(err)
		},
	)
}
