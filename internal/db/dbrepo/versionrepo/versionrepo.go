package versionrepo

import (
	"sort"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbmigrate"
	"github.com/oidc-mytoken/server/internal/model/version"
	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"
)

// SetVersionBefore sets that the before db migration commands for the passed version were executed
func SetVersionBefore(tx *sqlx.Tx, version string) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO version (version, bef) VALUES(?, current_timestamp()) ON DUPLICATE KEY UPDATE bef=current_timestamp()`, version)
		return err
	})
}

// SetVersionAfter sets that the after db migration commands for the passed version were executed
func SetVersionAfter(tx *sqlx.Tx, version string) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO version (version, aft) VALUES(?, current_timestamp()) ON DUPLICATE KEY UPDATE aft=current_timestamp()`, version)
		return err
	})
}

// UpdateTimes is a type for checking if the db migration commands for different mytoken version have been executed
type UpdateTimes struct {
	Version string
	Before  mysql.NullTime `db:"bef"`
	After   mysql.NullTime `db:"aft"`
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

// dbHasAllVersions checks that the database is compatible with the current version; assumes that DBVersionState is ordered
func (state DBVersionState) dBHasAllVersions() bool {
	for v, cmds := range dbmigrate.Migrate {
		if !state.dBHasVersion(v, cmds) {
			return false
		}
	}
	return true
}

// dbHasVersion checks that the database is compatible with the passed version; assumes that DBVersionState is ordered
func (state DBVersionState) dBHasVersion(v string, cmds dbmigrate.Commands) bool {
	i := sort.Search(len(state), func(i int) bool {
		return semver.Compare(state[i].Version, v) >= 0
	})
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
func GetVersionState(tx *sqlx.Tx) (state DBVersionState, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Select(&state, `SELECT version, bef, aft FROM version`)
	})
	if err != nil && strings.HasPrefix(err.Error(), "Error 1146: Table") { // Ignore table does not exist error
		err = nil
	}
	state.Sort()
	return
}

// ConnectToVersion connects to the mytoken database and asserts that the database is up-to-date to the current version
func ConnectToVersion() {
	db.Connect()
	state, err := GetVersionState(nil)
	if err != nil {
		log.WithError(err).Fatal()
	}
	if !state.dBHasAllVersions() {
		log.WithField("version", version.VERSION()).Fatal("database schema not updated to this server version")
	}
}
