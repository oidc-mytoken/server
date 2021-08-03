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

func getDoneMap(state versionrepo.DBVersionState, versions dbmigrate.VersionCommands) (map[string]bool, map[string]bool) {
	before := make(map[string]bool, len(versions))
	after := make(map[string]bool, len(versions))
	for v := range versions {
		before[v], after[v] = did(state, v)
	}
	return before, after
}

func migrateDB(mytokenNodes []string) error {
	v := version.VERSION()
	dbState, err := versionrepo.GetVersionState(nil)
	if err != nil {
		return err
	}
	return runUpdates(nil, dbState, mytokenNodes, v)
}

func runUpdates(tx *sqlx.Tx, dbState versionrepo.DBVersionState, mytokenNodes []string, version string) error {
	cmds := dbmigrate.Migrate
	beforeDone, afterDone := getDoneMap(dbState, cmds)
	if err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return runBeforeUpdates(tx, cmds, beforeDone)
	}); err != nil {
		return err
	}
	if !anyAfterUpdates(cmds, afterDone) { // If there are no after cmds to run, we are done
		return nil
	}
	waitUntilAllNodesOnVersion(mytokenNodes, version)

	return db.RunWithinTransaction(nil, func(tx *sqlx.Tx) error {
		return runAfterUpdates(tx, cmds, afterDone)
	})
}

func runBeforeUpdates(tx *sqlx.Tx, cmds dbmigrate.VersionCommands, beforeDone map[string]bool) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		for v, cs := range cmds {
			if err := updateCallback(tx, cs.Before, v, beforeDone, versionrepo.SetVersionBefore); err != nil {
				return err
			}
		}
		return nil
	})
}
func anyAfterUpdates(cmds dbmigrate.VersionCommands, afterDone map[string]bool) bool {
	for v, cs := range cmds {
		if len(cs.After) > 0 && !afterDone[v] {
			return true
		}
	}
	return false
}
func runAfterUpdates(tx *sqlx.Tx, cmds dbmigrate.VersionCommands, afterDone map[string]bool) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		for v, cs := range cmds {
			if err := updateCallback(tx, cs.After, v, afterDone, versionrepo.SetVersionAfter); err != nil {
				return err
			}
		}
		return nil
	})
}
func updateCallback(tx *sqlx.Tx, cmds []string, version string, done map[string]bool, dbUpdateCallback func(*sqlx.Tx, string) error) error {
	if len(cmds) == 0 {
		return nil
	}
	if done[version] {
		return nil
	}
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		if err := runDBCommands(tx, cmds); err != nil {
			return err
		}
		return dbUpdateCallback(tx, version)
	})
}

func waitUntilAllNodesOnVersion(mytokenNodes []string, version string) {
	allNodesOnVersion := len(mytokenNodes) == 0
	for !allNodesOnVersion {
		tmp := []string{}
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
	my, err := mytokenlib.NewMytokenProvider(node)
	if err != nil {
		return "", err
	}
	return my.Version, nil
}

func runDBCommands(tx *sqlx.Tx, cmds []string) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		for _, cmd := range cmds {
			cmd = strings.TrimSpace(cmd)
			if cmd != "" && !strings.HasPrefix(cmd, "--") {
				log.Trace(cmd)
				if _, err := tx.Exec(cmd); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
