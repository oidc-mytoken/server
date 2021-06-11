package dbmigrate

// Commands is a type for holding sql commands that should run before and after a version update
type Commands struct {
	Before []string `yaml:"before"`
	After  []string `yaml:"after"`
}

// VersionCommands is type holding the Commands that are related to a mytoken version
type VersionCommands map[string]Commands

// Migrate holds the VersionCommands for mytoken. These commands are used to migrate the database between mytoken versions.
var Migrate = VersionCommands{
	"0.2.0": {Before: v0_2_0_Before},
}
