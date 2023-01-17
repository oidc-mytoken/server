package main

import (
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	mytokenlib "github.com/oidc-mytoken/lib"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbmigrate"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/versionrepo"
	"github.com/oidc-mytoken/server/internal/model/version"
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

func migrateDB(mytokenNodes []string) error {
	v := "v" + version.VERSION
	dbState, err := versionrepo.GetVersionState(log.StandardLogger(), nil)
	if err != nil {
		return err
	}
	return runUpdates(dbState, mytokenNodes, v)
}

func runUpdates(dbState versionrepo.DBVersionState, mytokenNodes []string, version string) error {
	beforeDone, afterDone := getDoneMap(dbState)
	if err := runBeforeUpdates(beforeDone); err != nil {
		return err
	}
	if !anyAfterUpdates(afterDone) { // If there are no after cmds to run, we are done
		return nil
	}
	waitUntilAllNodesOnVersion(mytokenNodes, version)

	return runAfterUpdates(afterDone)
}

func runBeforeUpdates(beforeDone map[string]bool) error {
	for _, v := range dbmigrate.Versions {
		if err := updateCallback(
			dbmigrate.MigrationCommands[v].Before, v, beforeDone, versionrepo.SetVersionBefore,
		); err != nil {
			return err
		}
	}
	return nil
}
func anyAfterUpdates(afterDone map[string]bool) bool {
	for v, cs := range dbmigrate.MigrationCommands {
		if len(cs.After) > 0 && !afterDone[v] {
			return true
		}
	}
	return false
}
func runAfterUpdates(afterDone map[string]bool) error {
	for _, v := range dbmigrate.Versions {
		if err := updateCallback(
			dbmigrate.MigrationCommands[v].After, v, afterDone, versionrepo.SetVersionAfter,
		); err != nil {
			return err
		}
	}
	return nil
}
func updateCallback(
	cmds, version string, done map[string]bool,
	dbUpdateCallback func(log.Ext1FieldLogger, *sqlx.Tx, string) error,
) error {
	log.WithField("version", version).Info("Updating DB to version")
	if cmds == "" {
		return nil
	}
	if done[version] {
		log.WithField("version", version).Info("Skipping Update; DB already has this version.")
		return nil
	}
	if err := dbcl.RunDBCommands(cmds, dbConfig.DBConf, true); err != nil {
		return err
	}
	return db.Transact(
		log.StandardLogger(), func(tx *sqlx.Tx) error {
			return dbUpdateCallback(log.StandardLogger(), tx, version)
		},
	)
}

func waitUntilAllNodesOnVersion(mytokenNodes []string, version string) {
	allNodesOnVersion := len(mytokenNodes) == 0
	for !allNodesOnVersion {
		var tmp []string
		for _, n := range mytokenNodes {
			v, err := getVersionForNode(n)
			if err != nil {
				log.WithError(err).Error()
			}
			if v != version {
				tmp = append(tmp, n)
			}
		}
		mytokenNodes = tmp
		allNodesOnVersion = len(mytokenNodes) == 0
		time.Sleep(60 * time.Second)
	}
}

func getVersionForNode(node string) (string, error) {
	if !strings.HasPrefix(node, "http") {
		node = "https://" + node
	}
	my, err := mytokenlib.NewMytokenServer(node)
	if err != nil {
		return "", err
	}
	return my.ServerMetadata.Version, nil
}
