package dbmigrate

type Commands struct {
	Before []string `yaml:"before"`
	After  []string `yaml:"after"`
}

type VersionCommands map[string]Commands

var Migrate = VersionCommands{
	"0.2.0": {Before: v0_2_0_Before},
}
