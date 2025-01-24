package main

import (
	"crypto"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Songmu/prompter"
	"github.com/oidc-mytoken/utils/utils/fileutil"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
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

var askOverwriteFlag bool
var noOverwritePrompt bool
var guidedMode bool
var skipDB bool

var migrateDBConf struct {
	config.DBConf
	Hosts    cli.StringSlice
	force    bool
	confFile string
}

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
	Copyright:              "Karlsruhe Institute of Technology 2020-2025",
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
				&cli.Command{
					Name:        "ssh",
					Usage:       "Generates new ssh host key pairs",
					Description: "Generates new ssh host key paris to the location specified in the config file.",
					Action:      createSSHHostKeys,
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
				&cli.Command{
					Name:  "migrate",
					Usage: "Migrates the database to the latest version",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name: "nodes",
							Aliases: []string{
								"n",
								"s",
								"server",
							},
							Usage:       "The passed file lists the mytoken nodes / servers (one server per line)",
							EnvVars:     []string{"MYTOKEN_NODES_FILE"},
							TakesFile:   true,
							Placeholder: "FILE",
							Destination: &migrateDBConf.confFile,
						},
						&cli.BoolFlag{
							Name:    "force",
							Aliases: []string{"f"},
							Usage: "Force a complete database migration. It is not checked if mytoken servers are " +
								"compatible with the changes.",
							Destination:      &migrateDBConf.force,
							HideDefaultValue: true,
						},

						&cli.StringFlag{
							Name:        "db",
							Usage:       "The name of the database",
							EnvVars:     []string{"DB_DATABASE"},
							Value:       "mytoken",
							Destination: &migrateDBConf.DB,
							Placeholder: "DB",
						},
						&cli.StringFlag{
							Name:        "user",
							Aliases:     []string{"u"},
							Usage:       "The user for connecting to the database (Needs correct privileges)",
							EnvVars:     []string{"DB_USER"},
							Value:       "root",
							Destination: &migrateDBConf.User,
							Placeholder: "USER",
						},
						&cli.StringFlag{
							Name:    "password",
							Aliases: []string{"p"},
							Usage:   "The password for connecting to the database",
							EnvVars: []string{
								"DB_ROOT_PASSWORD",
								"DB_ROOT_PW",
							},
							Destination: &migrateDBConf.Password,
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
							Destination: &migrateDBConf.PasswordFile,
							Placeholder: "FILE",
						},
						&cli.StringSliceFlag{
							Name:    "host",
							Aliases: []string{"hosts"},
							Usage:   "The hostnames of the database nodes",
							EnvVars: []string{
								"DB_HOST",
								"DB_HOSTS",
								"DB_NODES",
							},
							Value:       cli.NewStringSlice("localhost"),
							Destination: &migrateDBConf.Hosts,
							Placeholder: "HOST",
							TakesFile:   true,
						},
					},
					Action: migrateDBAction,
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
	Action: guidedSetup,
	Flags: append(
		dbFlags,
		&cli.BoolFlag{
			Name: "ask-overwrites",
			Aliases: []string{
				"prompt-overwrites",
				"no-overwrite-skip",
			},
			Usage:       "If a file already exist, ask if it should be overwritten instead of skipping it.",
			Destination: &askOverwriteFlag,
		},
		&cli.BoolFlag{
			Name:        "skip-db",
			Usage:       "Skips db-related setups",
			Destination: &skipDB,
		},
	),
}

func main() {
	log.SetLevel(log.ErrorLevel)
	config.LoadForSetup()
	config.Get().Logging.Internal.Level = "error"
	loggerUtils.Init()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func guidedSetup(ctx *cli.Context) error {
	guidedMode = true
	noOverwritePrompt = !askOverwriteFlag
	fcs := []func(*cli.Context) error{
		installGEOIPDB,
		createMytokenSigningKey,
		createOIDCSigningKey,
		createFederationSigningKey,
		createSSHHostKeys,
	}
	if !skipDB {
		fcs = append(
			fcs,
			func(_ *cli.Context) error {
				fmt.Println("Setting up database...")
				return nil
			},
			createDB,
			createUser,
			func(_ *cli.Context) error {
				fmt.Println("Migrating database...")
				migrateDBConf.DBConf = rootDBCredentials.toDBConf()
				migrateDBConf.DBConf.DB = config.Get().DB.DB
				migrateDBConf.force = true
				db.ConnectConfig(migrateDBConf.DBConf)
				return migrateDB(nil)
			},
		)
	}

	for _, f := range fcs {
		if err := f(ctx); err != nil {
			return err
		}
		// fmt.Println()
	}
	fmt.Println("Setup done!")
	return nil
}

func installGEOIPDB(_ *cli.Context) error {
	f := config.Get().GeoIPDBFile
	if f == "" {
		fmt.Fprintln(os.Stderr, "No geoip database file specified")
		if !guidedMode {
			os.Exit(1)
		}
		return nil
	}
	if fileutil.FileExists(f) {
		fmt.Printf("Geo IP Database file '%s' already exists.\n", f)
		if noOverwritePrompt {
			return nil
		}
		if !prompter.YesNo(
			"Do you  want to re-download the geo ip database?", false,
		) {
			return nil
		}
	}
	archive, err := zipdownload.DownloadZipped("https://download.ip2location.com/lite/IP2LOCATION-LITE-DB1.IPV6.BIN.ZIP")
	if err != nil {
		return err
	}
	fmt.Println("Downloaded geo ip database")
	if err = os.WriteFile(config.Get().GeoIPDBFile, archive["IP2LOCATION-LITE-DB1.IPV6.BIN"], 0600); err != nil {
		return err
	}
	fmt.Printf("Geo IP Database file '%s' successfully installed.\n", f)
	return nil
}

func createMytokenSigningKey(_ *cli.Context) error {
	sk, _, err := jws.GenerateMytokenSigningKeyPair()
	if err != nil {
		return err
	}
	return writeSigningKey(sk, config.Get().Signing.Mytoken.KeyFile, "mytoken")
}
func createOIDCSigningKey(_ *cli.Context) error {
	sk, _, err := jws.GenerateOIDCSigningKeyPair()
	if err != nil {
		return err
	}
	return writeSigningKey(sk, config.Get().Signing.OIDC.KeyFile, "oidc")
}
func createFederationSigningKey(_ *cli.Context) error {
	if !config.Get().Features.Federation.Enabled && guidedMode {
		return nil
	}
	sk, _, err := jws.GenerateFederationSigningKeyPair()
	if err != nil {
		return err
	}
	return writeSigningKey(sk, config.Get().Features.Federation.Signing.KeyFile, "openid federations")
}

func writeSigningKey(sk crypto.Signer, keyPathFromConfig, purpose string) error {
	str := jws.ExportPrivateKeyAsPemStr(sk)
	if sigKeyFile == "" || guidedMode {
		sigKeyFile = keyPathFromConfig
	}
	if sigKeyFile == "" {
		if guidedMode && purpose != "mytoken" {
			return nil
		}
		return errors.New("no signing key file specified")
	}
	if fileutil.FileExists(sigKeyFile) {
		fmt.Printf("%s signing key file '%s' already exists.\n", purpose, sigKeyFile)
		if noOverwritePrompt {
			return nil
		}
		if !prompter.YesNo("Do you  want to overwrite it?", false) {
			return nil
		}
	}
	if err := mkdir(filepath.Dir(sigKeyFile)); err != nil {
		return err
	}
	if err := os.WriteFile(sigKeyFile, []byte(str), 0600); err != nil {
		return err
	}
	fmt.Printf("Wrote %s signing key to file '%s'.\n", purpose, sigKeyFile)
	return nil
}

func createSSHHostKeys(_ *cli.Context) error {
	if guidedMode && !config.Get().Features.SSH.Enabled {
		return nil
	}
	for _, keyFile := range config.Get().Features.SSH.KeyFiles {
		if fileutil.FileExists(keyFile) {
			fmt.Printf("ssh host key file '%s' already exists.\n", keyFile)
			if noOverwritePrompt || !prompter.YesNo(
				"Do you  want to overwrite it?", false,
			) {
				continue
			}
		}
		if err := mkdir(filepath.Dir(keyFile)); err != nil {
			return err
		}
		var typ string
		var additionalArgs []string
		switch {
		case strings.HasSuffix(keyFile, "_rsa_key"):
			typ = "rsa"
			additionalArgs = append(additionalArgs, "-b", "4096")
		case strings.HasSuffix(keyFile, "_ed25519_key"):
			typ = "ed25519"
		case strings.HasSuffix(keyFile, "_ecdsa_key"):
			typ = "ecdsa"
			additionalArgs = append(additionalArgs, "-b", "521")
		}
		keygenCmd := fmt.Sprintf(`ssh-keygen -t %s %s -f %s -N ""`, typ, strings.Join(additionalArgs, " "), keyFile)
		cmd := exec.Command("sh", "-c", keygenCmd)
		if err := cmd.Run(); err != nil {
			log.WithField("filepath", keyFile).WithError(err).Error("Failed to generate key")
			return err
		} else {
			fmt.Printf("Generated ssh host key '%s'.\n", keyFile)
		}
	}
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
func _getSetVarsAfterExec() (string, error) {
	return readSQLFile("scripts/vars_after.sql")
}
func getSetVarsCommands(db, user, password string) (string, error) {
	cmds, err := _getSetVars()
	if err != nil {
		return "", err
	}
	if db != "" {
		cmds += "\n" + fmt.Sprintf(`EXECUTE setDB USING '%s';`, db)
	}
	if user != "" {
		cmds += "\n" + fmt.Sprintf(`EXECUTE setUser USING '%s';`, user)
	}
	if password != "" {
		cmds += "\n" + fmt.Sprintf(`EXECUTE setPassword USING '%s';`, password)
	}
	after, err := _getSetVarsAfterExec()
	if err != nil {
		return "", err
	}
	cmds += after
	return cmds, nil
}
func getDBCmds() (string, error) {
	return readSQLFile("scripts/db.sql")
}
func getUserCmds() (string, error) {
	return readSQLFile("scripts/user.sql")
}

func dbEnsureRootPW() {
	dbconf := rootDBCredentials.toDBConf()
	user := rootDBCredentials.User
	if user == "" {
		user = "root"
	}
	if dbconf.GetPassword() == "" {
		rootDBCredentials.Password = prompter.Password(
			fmt.Sprintf(
				"Enter db password for user '%s'", user,
			),
		)
	}
}

func createDB(_ *cli.Context) error {
	dbEnsureRootPW()
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
	dbEnsureRootPW()
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

func migrateDBAction(context *cli.Context) error {
	var mytokenNodes []string
	if context.Args().Len() > 0 {
		mytokenNodes = context.Args().Slice()
	} else if migrateDBConf.confFile != "" {
		data := string(fileutil.MustReadFile(migrateDBConf.confFile))
		mytokenNodes = strings.Split(data, "\n")
	} else if os.Getenv("MYTOKEN_NODES") != "" {
		mytokenNodes = strings.Split(os.Getenv("MYTOKEN_NODES"), ",")
	} else if !migrateDBConf.force {
		fmt.Fprintln(
			os.Stderr,
			"No mytoken servers specified. Please provide mytoken servers or use '-f' to "+
				"force database migration.",
		)
		os.Exit(1)
	}
	if migrateDBConf.GetPassword() == "" {
		migrateDBConf.Password = prompter.Password(
			fmt.Sprintf(
				"Enter db password for user '%s'", migrateDBConf.User,
			),
		)
	}
	migrateDBConf.ReconnectInterval = 60
	migrateDBConf.DBConf.Hosts = migrateDBConf.Hosts.Value()
	tmpScheduleEnabled := migrateDBConf.DBConf.EnableScheduledCleanup
	migrateDBConf.DBConf.EnableScheduledCleanup = false
	db.ConnectConfig(migrateDBConf.DBConf)
	migrateDBConf.DBConf.EnableScheduledCleanup = tmpScheduleEnabled
	return migrateDB(mytokenNodes)
}

func mkdir(path string) error {
	return os.MkdirAll(path, 0750)
}
