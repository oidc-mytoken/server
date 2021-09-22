package main

import (
	"fmt"
	"os"
	"os/exec"
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

func getDoneMap(state versionrepo.DBVersionState) (map[string]bool, map[string]bool) {
	before := make(map[string]bool, len(dbmigrate.Versions))
	after := make(map[string]bool, len(dbmigrate.Versions))
	for _, v := range dbmigrate.Versions {
		before[v], after[v] = did(state, v)
	}
	return before, after
}

func migrateDB(mytokenNodes []string) error {
	v := "v" + version.VERSION()
	dbState, err := versionrepo.GetVersionState(nil)
	if err != nil {
		return err
	}
	return runUpdates(nil, dbState, mytokenNodes, v)
}

func runUpdates(tx *sqlx.Tx, dbState versionrepo.DBVersionState, mytokenNodes []string, version string) error {
	beforeDone, afterDone := getDoneMap(dbState)
	if err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return runBeforeUpdates(tx, beforeDone)
	}); err != nil {
		return err
	}
	if !anyAfterUpdates(afterDone) { // If there are no after cmds to run, we are done
		return nil
	}
	waitUntilAllNodesOnVersion(mytokenNodes, version)

	return db.RunWithinTransaction(nil, func(tx *sqlx.Tx) error {
		return runAfterUpdates(tx, afterDone)
	})
}

func runBeforeUpdates(tx *sqlx.Tx, beforeDone map[string]bool) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		for _, v := range dbmigrate.Versions {
			if err := updateCallback(tx, dbmigrate.MigrationCommands[v].Before, v, beforeDone, versionrepo.SetVersionBefore); err != nil {
				return err
			}
		}
		return nil
	})
}
func anyAfterUpdates(afterDone map[string]bool) bool {
	for v, cs := range dbmigrate.MigrationCommands {
		if len(cs.After) > 0 && !afterDone[v] {
			return true
		}
	}
	return false
}
func runAfterUpdates(tx *sqlx.Tx, afterDone map[string]bool) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		for _, v := range dbmigrate.Versions {
			if err := updateCallback(tx, dbmigrate.MigrationCommands[v].After, v, afterDone, versionrepo.SetVersionAfter); err != nil {
				return err
			}
		}
		return nil
	})
}
func updateCallback(tx *sqlx.Tx, cmds, version string, done map[string]bool, dbUpdateCallback func(*sqlx.Tx, string) error) error {
	log.WithField("version", version).Info("Updating DB to version")
	if len(cmds) == 0 {
		return nil
	}
	if done[version] {
		log.WithField("version", version).Info("Skipping Update; DB already has this version.")
		return nil
	}
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		if err := runDBCommands(cmds); err != nil {
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

func runDBCommands(cmds string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("mysql -uroot -p%s --protocol tcp -h %s %s", dbConfig.GetPassword(), dbConfig.Hosts.Value()[0], dbConfig.DB))
	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if _, err = cmdIn.Write([]byte(cmds)); err != nil {
		return err
	}
	if err = cmdIn.Close(); err != nil {
		return err
	}
	return cmd.Run()
}
