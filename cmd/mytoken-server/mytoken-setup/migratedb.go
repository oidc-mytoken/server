package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbmigrate"
	"github.com/oidc-mytoken/server/internal/db/dbmigrate/gomigrations"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/versionrepo"
	"github.com/oidc-mytoken/server/internal/utils/dbcl"
)

func did(state versionrepo.DBVersionState, version string) (beforeDone, afterDone bool) {
	for _, entry := range state {
		if entry.Version == version {
			if entry.Before.Valid {
				beforeDone = true
			}
			if entry.After.Valid {
				afterDone = true
			}
			return
		}
	}
	return
}

func getDoneMap(state versionrepo.DBVersionState) (map[string]bool, map[string]bool) {
	before := make(map[string]bool, len(dbmigrate.Versions))
	after := make(map[string]bool, len(dbmigrate.Versions))
	for _, v := range dbmigrate.Versions {
		before[v], after[v] = did(state, v)
	}
	return before, after
}

func migrateDB() error {
	dbState, err := versionrepo.GetVersionState(log.StandardLogger(), nil)
	if err != nil {
		return err
	}
	return runUpdates(dbState)
}

func runUpdates(dbState versionrepo.DBVersionState) error {
	beforeDone, afterDone := getDoneMap(dbState)
	for _, v := range dbmigrate.Versions {
		if err := updateCallback(
			dbmigrate.MigrationCommands[v].Before, v, "pre", beforeDone, versionrepo.SetVersionBefore,
		); err != nil {
			return err
		}
		if err := gomigrations.RunGoMigration(nil, v); err != nil {
			return err
		}
		if err := updateCallback(
			dbmigrate.MigrationCommands[v].After, v, "post", afterDone, versionrepo.SetVersionAfter,
		); err != nil {
			return err
		}
	}
	return nil
}

func updateCallback(
	cmds, version, code string, done map[string]bool,
	dbUpdateCallback func(log.Ext1FieldLogger, *sqlx.Tx, string) error,
) error {
	fmt.Printf("Updating DB to version %s %s\n", version, code)
	if cmds == "" {
		return nil
	}
	if done[version] {
		fmt.Printf("Skipping Update; DB already has version %s %s\n", version, code)
		return nil
	}
	if err := dbcl.RunDBCommands(cmds, migrateDBConf.DBConf, true); err != nil {
		return err
	}
	return db.Transact(
		log.StandardLogger(), func(tx *sqlx.Tx) error {
			return dbUpdateCallback(log.StandardLogger(), tx, version)
		},
	)
}
