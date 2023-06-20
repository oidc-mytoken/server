package main

import (
	"crypto"
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/Songmu/prompter"
	"github.com/oidc-mytoken/utils/utils/fileutil"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/jws"
	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/utils/dbcl"
	loggerUtils "github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/internal/utils/zipdownload"
)

type _rootDBCredentials struct {
	User         string
	Password     string
	PasswordFile string
}

var rootDBCredentials _rootDBCredentials

func (cred _rootDBCredentials) toDBConf() config.DBConf {
	return config.DBConf{
		Hosts:             config.Get().DB.Hosts,
		User:              cred.User,
		Password:          cred.Password,
		PasswordFile:      cred.PasswordFile,
		ReconnectInterval: config.Get().DB.ReconnectInterval,
	}
}

var dbFlags = []cli.Flag{
	&cli.StringFlag{
		Name: "user",
		Aliases: []string{
			"u",
			"root-user",
			"db-user",
		},
		Usage:       "The username for the (root) user used for setting up the db",
		EnvVars:     []string{"DB_USER"},
		Value:       "root",
		Destination: &rootDBCredentials.User,
		Placeholder: "ROOT",
	},
	&cli.StringFlag{
		Name: "password",
		Aliases: []string{
			"p",
			"pw",
			"db-password",
			"db-pw",
		},
		Usage: "The password for the (root) user used for setting up the db",
		EnvVars: []string{
			"DB_PW",
			"DB_PASSWORD",
		},
		Destination: &rootDBCredentials.Password,
		Placeholder: "PASSWORD",
	},
	&cli.StringFlag{
		Name:    "password-file",
		Aliases: []string{"pw-file"},
		Usage:   "Read the password for connecting to the database from this file",
		EnvVars: []string{
			"DB_PASSWORD_FILE",
			"DB_PW_FILE",
		},
		Destination: &rootDBCredentials.PasswordFile,
		TakesFile:   true,
		Placeholder: "FILE",
	},
}

var sigKeyFile string

var sigKeyFlag = &cli.StringFlag{
	Name: "key-file",
	Aliases: []string{
		"file",
		"f",
		"out",
		"o",
	},
	Usage: "Write the signing key to this file, " +
		"instead of the one configured in the config file",
	EnvVars: []string{
		"KEY_FILE",
		"SIGNING_KEY",
	},
	Destination: &sigKeyFile,
	TakesFile:   true,
	Placeholder: "FILE",
}

var app = &cli.App{
	Name:     "mytoken-setup",
	Usage:    "Command line client for easily setting up a mytoken server",
	Version:  version.VERSION,
	Compiled: time.Time{},
	Authors: []*cli.Author{
		{
			Name:  "Gabriel Zachmann",
			Email: "gabriel.zachmann@kit.edu",
		},
	},
	Copyright:              "Karlsruhe Institute of Technology 2020-2022",
	UseShortOptionHandling: true,
	Commands: cli.Commands{
		&cli.Command{
			Name: "signing-key",
			Aliases: []string{
				"key",
				"keys",
				"signing-keys",
			},
			Subcommands: cli.Commands{
				&cli.Command{
					Name: "mytoken",
					Aliases: []string{
						"mt",
						"MT",
					},
					Flags: []cli.Flag{
						sigKeyFlag,
					},
					Usage:       "Generates a new mytoken signing key",
					Description: "Generates a new mytoken signing key according to the properties specified in the config file and stores it.",
					Action:      createMytokenSigningKey,
				},
				&cli.Command{
					Name:    "oidc",
					Aliases: []string{"OIDC"},
					Flags: []cli.Flag{
						sigKeyFlag,
					},
					Usage:       "Generates a new oidc signing key",
					Description: "Generates a new oidc signing key according to the properties specified in the config file and stores it.",
					Action:      createOIDCSigningKey,
				},
			},
			Flags: []cli.Flag{
				sigKeyFlag,
			},
		},
		&cli.Command{
			Name:  "install",
			Usage: "Installs needed dependencies",
			Subcommands: cli.Commands{
				&cli.Command{
					Name:    "geoip-db",
					Aliases: []string{"geo-ip-db"},
					Usage:   "Installs the ip geolocation database.",
					Action:  installGEOIPDB,
				},
			},
		},
		&cli.Command{
			Name:  "db",
			Usage: "Setups for the database",
			Flags: append([]cli.Flag{}, dbFlags...),
			Subcommands: cli.Commands{
				&cli.Command{
					Name:    "db",
					Aliases: []string{"database"},
					Usage:   "Creates the database in the database server",
					Action:  createDB,
					Flags:   append([]cli.Flag{}, dbFlags...),
				},
				&cli.Command{
					Name:   "user",
					Usage:  "Creates the normal database user",
					Action: createUser,
					Flags:  append([]cli.Flag{}, dbFlags...),
				},
			},
		},
		&cli.Command{
			Name:  "federation",
			Usage: "Setups for federations",
			Subcommands: cli.Commands{
				&cli.Command{
					Name:    "signing-key",
					Aliases: []string{"key"},
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name: "key-file",
							Aliases: []string{
								"file",
								"f",
								"out",
								"o",
							},
							Usage: "Write the signing key to this file, " +
								"instead of the one configured in the config file",
							EnvVars: []string{
								"KEY_FILE",
								"SIGNING_KEY",
							},
							Destination: &sigKeyFile,
							TakesFile:   true,
							Placeholder: "FILE",
						},
					},
					Usage:       "Generates a new signing key",
					Description: "Generates a new signing key according to the properties specified in the config file and stores it.",
					Action:      createFederationSigningKey,
				},
			},
		},
	},
}

func main() {
	config.LoadForSetup()
	loggerUtils.Init()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func installGEOIPDB(_ *cli.Context) error {
	archive, err := zipdownload.DownloadZipped("https://download.ip2location.com/lite/IP2LOCATION-LITE-DB1.IPV6.BIN.ZIP")
	if err != nil {
		return err
	}
	log.Debug("Downloaded zip file")
	err = os.WriteFile(config.Get().GeoIPDBFile, archive["IP2LOCATION-LITE-DB1.IPV6.BIN"], 0600)
	if err == nil {
		log.WithField("file", config.Get().GeoIPDBFile).Debug("Installed geo ip database")
		fmt.Printf("Installed geo ip database file to '%s'.\n", config.Get().GeoIPDBFile)
	}
	return err
}

func createMytokenSigningKey(_ *cli.Context) error {
	sk, _, err := jws.GenerateMytokenSigningKeyPair()
	if err != nil {
		return err
	}
	return writeSigningKey(sk, config.Get().Signing.Mytoken.KeyFile)
}
func createOIDCSigningKey(_ *cli.Context) error {
	sk, _, err := jws.GenerateOIDCSigningKeyPair()
	if err != nil {
		return err
	}
	return writeSigningKey(sk, config.Get().Signing.OIDC.KeyFile)
}
func createFederationSigningKey(_ *cli.Context) error {
	sk, _, err := jws.GenerateFederationSigningKeyPair()
	if err != nil {
		return err
	}
	return writeSigningKey(sk, config.Get().Features.Federation.Signing.KeyFile)
}

func writeSigningKey(sk crypto.Signer, keyFileFromConfig string) error {
	str := jws.ExportPrivateKeyAsPemStr(sk)
	if sigKeyFile == "" {
		sigKeyFile = keyFileFromConfig
	}
	if fileutil.FileExists(sigKeyFile) {
		log.WithField("filepath", sigKeyFile).Debug("File already exists")
		if !prompter.YesNo(fmt.Sprintf("File '%s' already exists. Do you  want to overwrite it?", sigKeyFile), false) {
			os.Exit(1)
		}
	}
	if err := os.WriteFile(sigKeyFile, []byte(str), 0600); err != nil {
		return err
	}
	log.WithField("filepath", sigKeyFile).Debug("Wrote key to file")
	fmt.Printf("Wrote key to file '%s'.\n", sigKeyFile)
	return nil
}

//go:embed scripts
var sqlScripts embed.FS

func readSQLFile(path string) (string, error) {
	data, err := sqlScripts.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func _getSetVars() (string, error) {
	return readSQLFile("scripts/vars.sql")
}
func getSetVarsCommands(db, user, password string) (string, error) {
	cmds, err := _getSetVars()
	if err != nil {
		return "", err
	}
	if db != "" {
		cmds += fmt.Sprintf(`EXECUTE setDB USING '%s';\n`, db)
	}
	if user != "" {
		cmds += fmt.Sprintf(`EXECUTE setUser USING '%s';\n`, user)
	}
	if password != "" {
		cmds += fmt.Sprintf(`EXECUTE setPassword USING '%s';\n`, password)
	}
	return cmds, nil
}
func getDBCmds() (string, error) {
	return readSQLFile("scripts/db.sql")
}
func getUserCmds() (string, error) {
	return readSQLFile("scripts/user.sql")
}

func createDB(_ *cli.Context) error {
	cmds, err := getSetVarsCommands(config.Get().DB.DB, config.Get().DB.User, config.Get().DB.GetPassword())
	if err != nil {
		return err
	}
	dbCmds, err := getDBCmds()
	if err != nil {
		return err
	}
	cmds += dbCmds
	return dbcl.RunDBCommands(cmds, rootDBCredentials.toDBConf(), true)
}

func createUser(_ *cli.Context) error {
	cmds, err := getSetVarsCommands(config.Get().DB.DB, config.Get().DB.User, config.Get().DB.GetPassword())
	if err != nil {
		return err
	}
	userCmds, err := getUserCmds()
	if err != nil {
		return err
	}
	cmds += userCmds
	return dbcl.RunDBCommands(cmds, rootDBCredentials.toDBConf(), true)
}
