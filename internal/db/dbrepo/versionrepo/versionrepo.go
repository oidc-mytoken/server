package versionrepo

import (
	"database/sql"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbmigrate"
	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
)

// SetVersionBefore sets that the before db migration commands for the passed version were executed
func SetVersionBefore(rlog log.Ext1FieldLogger, tx *sqlx.Tx, version string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Version_SetBefore(?)`, version)
			return errors.WithStack(err)
		},
	)
}

// SetVersionAfter sets that the after db migration commands for the passed version were executed
func SetVersionAfter(rlog log.Ext1FieldLogger, tx *sqlx.Tx, version string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Version_SetAfter(?)`, version)
			return errors.WithStack(err)
		},
	)
}

// UpdateTimes is a type for checking if the db migration commands for different mytoken version have been executed
type UpdateTimes struct {
	Version string
	Before  sql.NullTime `db:"bef"`
	After   sql.NullTime `db:"aft"`
}

// DBVersionState describes the version state of the db
type DBVersionState []UpdateTimes

// Len returns the len of DBVersionState
func (state DBVersionState) Len() int { return len(state) }

// Swap swaps to elements of DBVersionState
func (state DBVersionState) Swap(i, j int) { state[i], state[j] = state[j], state[i] }

// Less checks if a version is less than another
func (state DBVersionState) Less(i, j int) bool {
	a, b := state[i].Version, state[j].Version
	return semver.Compare(a, b) < 0
}

// Sort sorts this DBVersionState by the version
func (state DBVersionState) Sort() {
	sort.Sort(state)
}

// dbHasAllVersions checks that the database is compatible with the current version; assumes that DBVersionState is
// ordered
func (state DBVersionState) dBHasAllVersions() (hasAllVersions bool, missingVersions []string) {
	for v, cmds := range dbmigrate.MigrationCommands {
		if !state.dBHasVersion(v, cmds) {
			missingVersions = append(missingVersions, v)
		}
	}
	hasAllVersions = len(missingVersions) == 0
	return
}

// dbHasVersion checks that the database is compatible with the passed version; assumes that DBVersionState is ordered
func (state DBVersionState) dBHasVersion(v string, cmds dbmigrate.Commands) bool {
	i := sort.Search(
		len(state), func(i int) bool {
			return semver.Compare(state[i].Version, v) >= 0
		},
	)
	if i >= len(state) || state[i].Version != v { // we have to check that i really points to v
		return false
	}
	ok := true
	if len(cmds.Before) > 0 {
		ok = ok && state[i].Before.Valid
	}
	if len(cmds.After) > 0 {
		ok = ok && state[i].After.Valid
	}
	return ok
}

// GetVersionState returns the DBVersionState
func GetVersionState(rlog log.Ext1FieldLogger, tx *sqlx.Tx) (state DBVersionState, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			err = errors.WithStack(tx.Select(&state, `CALL Version_Get()`))
			if err == nil {
				return nil
			}
			if !strings.HasSuffix(err.Error(), "Version_Get does not exist") {
				return err
			}
			return errors.WithStack(tx.Select(&state, `SELECT version, bef, aft FROM version`))
		},
	)
	if err != nil && strings.HasPrefix(errorfmt.Error(err), "Error 1146: Table") { // Ignore table does not exist error
		err = nil
	}
	state.Sort()
	return
}

// ConnectToVersion connects to the mytoken database and asserts that the database is up-to-date to the current version
func ConnectToVersion() {
	db.Connect()
	state, err := GetVersionState(log.StandardLogger(), nil)
	if err != nil {
		log.WithError(err).Fatal()
	}
	if hasAllVersions, missingVersions := state.dBHasAllVersions(); !hasAllVersions {
		log.WithFields(
			log.Fields{
				"server_version":         version.VERSION(),
				"missing_versions_in_db": missingVersions,
			},
		).Fatal("database schema not updated to this server version")
	}
}
