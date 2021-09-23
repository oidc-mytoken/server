package dbmigrate

import (
	"embed"
	"fmt"
	"io/fs"

	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"

	"github.com/oidc-mytoken/server/shared/utils"
)

// Commands is a type for holding sql commands that should run before and after a version update
type Commands struct {
	Before string `yaml:"before"`
	After  string `yaml:"after"`
}

// VersionCommands is type holding the Commands that are related to a mytoken version
type VersionCommands map[string]Commands

// MigrationCommands holds the VersionCommands for mytoken. These commands are used to migrate the database between mytoken
// versions.
var MigrationCommands = VersionCommands{}

// Versions holds all versions for which migration commands are available
var Versions []string

//go:embed scripts
var migrationScripts embed.FS

func init() {
	Versions = []string{}
	if err := fs.WalkDir(fs.FS(migrationScripts), ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		Versions = append(Versions, utils.RSplitN(name, ".", 3)[0])
		return nil
	}); err != nil {
		log.WithError(err).Fatal()
	}
	semver.Sort(Versions)
	for _, v := range Versions {
		MigrationCommands[v] = Commands{
			Before: readBeforeFile(v),
			After:  readAfterFile(v),
		}
	}
}

func readBeforeFile(version string) string {
	return _readSQLFile(version, "pre")
}

func readAfterFile(version string) string {
	return _readSQLFile(version, "post")
}

func _readSQLFile(version, typeString string) string {
	data, err := migrationScripts.ReadFile(fmt.Sprintf("scripts/%s.%s.sql", version, typeString))
	if err != nil {
		return ""
	}
	return string(data)
}
