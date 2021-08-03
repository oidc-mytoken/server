package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/shared/utils/fileutil"
	"github.com/zachmann/cli/v2"
	"golang.org/x/term"

	"github.com/oidc-mytoken/server/internal/model/version"
	log "github.com/sirupsen/logrus"
)

var configFile string
var force bool

var dbConfig struct {
	config.DBConf
	Hosts cli.StringSlice
}

var app = &cli.App{
	Name:     "mytoken-migratedb",
	Usage:    "Command line client for easy database migration between mytoken versions",
	Version:  version.VERSION(),
	Compiled: time.Time{},
	Authors: []*cli.Author{{
		Name:  "Gabriel Zachmann",
		Email: "gabriel.zachmann@kit.edu",
	}},
	Copyright:              "Karlsruhe Institute of Technology 2020-2021",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "nodes",
			Aliases:     []string{"n", "s", "server"},
			Usage:       "The passed file lists the mytoken nodes / servers (one server per line)",
			EnvVars:     []string{"MYTOKEN_NODES_FILE"},
			TakesFile:   true,
			Placeholder: "FILE",
			Destination: &configFile,
		},
		&cli.BoolFlag{
			Name:             "force",
			Aliases:          []string{"f"},
			Usage:            "Force a complete database migration. Mytoken servers are not checked if they are compatible with the changes.",
			Destination:      &force,
			HideDefaultValue: true,
		},

		&cli.StringFlag{
			Name:        "db",
			Usage:       "The name of the database",
			EnvVars:     []string{"DB_DATABASE"},
			Value:       "mytoken",
			Destination: &dbConfig.DB,
			Placeholder: "DB",
		},
		&cli.StringFlag{
			Name:        "user",
			Aliases:     []string{"u"},
			Usage:       "The user for connecting to the database (Needs correct privileges)",
			EnvVars:     []string{"DB_USER"},
			Value:       "root",
			Destination: &dbConfig.User,
			Placeholder: "USER",
		},
		&cli.StringFlag{
			Name:        "password",
			Aliases:     []string{"p"},
			Usage:       "The password for connecting to the database",
			EnvVars:     []string{"DB_PASSWORD"},
			Destination: &dbConfig.Password,
			Placeholder: "PASSWORD",
		},
		&cli.StringFlag{
			Name:        "password-file",
			Aliases:     []string{"pass-file"},
			Usage:       "Read the password for connecting to the database from this file",
			EnvVars:     []string{"DB_PASSWORD_FILE"},
			Destination: &dbConfig.PasswordFile,
			Placeholder: "FILE",
		},
		&cli.StringSliceFlag{
			Name:        "host",
			Aliases:     []string{"hosts"},
			Usage:       "The hostnames of the database nodes",
			EnvVars:     []string{"DB_HOST", "DB_HOSTS", "DB_NODES"},
			Value:       cli.NewStringSlice("localhost"),
			Destination: &dbConfig.Hosts,
			Placeholder: "HOST",
		},
	},
	Action: func(context *cli.Context) error {
		mytokenNodes := []string{}
		if context.Args().Len() > 0 {
			mytokenNodes = context.Args().Slice()
		} else if configFile != "" {
			mytokenNodes = readConfigFile(configFile)
		} else if os.Getenv("MYTOKEN_NODES") != "" {
			mytokenNodes = strings.Split(os.Getenv("MYTOKEN_NODES"), ",")
		} else if !force {
			return fmt.Errorf("No mytoken servers specified. Please provide mytoken servers or use '-f' to force database migration.")
		}
		dbConfig.ReconnectInterval = 60
		dbConfig.DBConf.Hosts = dbConfig.Hosts.Value()
		db.ConnectConfig(dbConfig.DBConf)
		return migrateDB(nil, mytokenNodes)
	},
}

func readConfigFile(file string) []string {
	data := string(fileutil.MustReadFile(file))
	return strings.Split(data, "\n")
}

func main() {

	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil {
		cli.HelpWrapAt = termWidth
	}

	if err = app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
