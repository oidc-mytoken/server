package profilerepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
)

// DeleteProfile deletes a profile for the passed group and id
func DeleteProfile(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string, id uuid.UUID) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Profiles_DeleteProfiles(?,?)`, id, group)
			return errors.WithStack(err)
		},
	)
}

// DeleteRestrictions deletes a restrictions template for the passed group and id
func DeleteRestrictions(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string, id uuid.UUID) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Profiles_DeleteRestrictions(?,?)`, id, group)
			return errors.WithStack(err)
		},
	)
}

// DeleteRotation deletes a rotation template for the passed group and id
func DeleteRotation(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string, id uuid.UUID) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Profiles_DeleteRotations(?,?)`, id, group)
			return errors.WithStack(err)
		},
	)
}

// DeleteCapabilities deletes a capability template for the passed group and id
func DeleteCapabilities(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string, id uuid.UUID) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Profiles_DeleteCapabilities(?,?)`, id, group)
			return errors.WithStack(err)
		},
	)
}
