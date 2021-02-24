package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/jessevdk/go-flags"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db/cluster"
	"github.com/oidc-mytoken/server/internal/db/dbdefinition"
	"github.com/oidc-mytoken/server/internal/jws"
	"github.com/oidc-mytoken/server/internal/model"
	loggerUtils "github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/internal/utils/zipdownload"
	model2 "github.com/oidc-mytoken/server/pkg/model"
	event "github.com/oidc-mytoken/server/shared/supertoken/event/pkg"
	"github.com/oidc-mytoken/server/shared/utils/fileutil"
)

var genSigningKeyComm commandGenSigningKey
var createDBComm commandCreateDB
var installComm struct {
	GeoIP commandInstallGeoIPDB `command:"geoip-db" description:"Installs the ip geolocation database."`
}

func main() {
	config.LoadForSetup()
	loggerUtils.Init()

	parser := flags.NewNamedParser("mytoken", flags.HelpFlag|flags.PassDoubleDash)
	parser.AddCommand("signing-key", "Generates a new signing key", "Generates a new signing key according to the properties specified in the config file and stores it.", &genSigningKeyComm)
	parser.AddCommand("db", "Setups the database", "Setups the database as needed and specified in the config file.", &createDBComm)
	parser.AddCommand("install", "Installs needed dependencies", "", &installComm)
	_, err := parser.Parse()
	if err != nil {
		var flagError *flags.Error
		if errors.As(err, &flagError) {
			if flagError.Type == flags.ErrHelp {
				fmt.Println(err)
				os.Exit(0)
			}
		}
		log.WithError(err).Fatal()
		os.Exit(1)
	}

}

type commandGenSigningKey struct{}
type commandCreateDB struct {
	Username string  `short:"u" long:"user" default:"root" description:"This username is used to connect to the database to create a new database, database user, and tables."`
	Password *string `short:"p" optional:"true" optional-value:"" long:"password" description:"The password for the database user"`
}
type commandInstallGeoIPDB struct{}

// Execute implements the flags.Commander interface
func (c *commandInstallGeoIPDB) Execute(args []string) error {
	archive, err := zipdownload.DownloadZipped("https://download.ip2location.com/lite/IP2LOCATION-LITE-DB1.IPV6.BIN.ZIP")
	if err != nil {
		return err
	}
	log.Debug("Downloaded zip file")
	err = ioutil.WriteFile(config.Get().GeoIPDBFile, archive["IP2LOCATION-LITE-DB1.IPV6.BIN"], 0644)
	if err == nil {
		log.WithField("file", config.Get().GeoIPDBFile).Debug("Installed geo ip database")
		fmt.Printf("Installed geo ip database file to '%s'.\n", config.Get().GeoIPDBFile)
	}
	return err
}

// Execute implements the flags.Commander interface
func (c *commandGenSigningKey) Execute(args []string) error {
	sk, _, err := jws.GenerateKeyPair()
	if err != nil {
		return err
	}
	str := jws.ExportPrivateKeyAsPemStr(sk)
	filepath := config.Get().Signing.KeyFile
	if fileutil.FileExists(filepath) {
		log.WithField("filepath", filepath).Debug("File already exists")
		if !prompter.YesNo(fmt.Sprintf("File '%s' already exists. Do you  want to overwrite it?", filepath), false) {
			os.Exit(1)
		}
	}
	if err = ioutil.WriteFile(filepath, []byte(str), 0600); err != nil {
		return err
	}
	log.WithField("filepath", filepath).Debug("Wrote key to file")
	fmt.Printf("Wrote key to file '%s'.\n", filepath)
	return nil
}

// Execute implements the flags.Commander interface
func (c *commandCreateDB) Execute(args []string) error {
	password := ""
	if c.Password != nil && *c.Password == "" { // -p specified without argument
		password = prompter.Password("Database Password")
	}
	db := cluster.NewFromConfig(config.DBConf{
		Hosts:    config.Get().DB.Hosts,
		User:     c.Username,
		Password: password,
	})
	if err := checkDB(db); err != nil {
		return err
	}
	err := db.Transact(func(tx *sqlx.Tx) error {
		if err := createDB(tx); err != nil {
			return err
		}
		if err := createUser(tx); err != nil {
			return err
		}
		if err := createTables(tx); err != nil {
			return err
		}
		if err := addPredefinedValues(tx); err != nil { // skipcq RVV-B0005
			return err
		}
		return nil
	})
	if err == nil {
		fmt.Println("Prepared database.")
	}
	return err
}

func addPredefinedValues(tx *sqlx.Tx) error {
	for _, attr := range model.Attributes {
		if _, err := tx.Exec(`INSERT IGNORE INTO Attributes (attribute) VALUES(?)`, attr); err != nil {
			return err
		}
	}
	log.WithField("database", config.Get().DB.DB).Debug("Added attribute values")
	for _, evt := range event.AllEvents {
		if _, err := tx.Exec(`INSERT IGNORE INTO Events (event) VALUES(?)`, evt); err != nil {
			return err
		}
	}
	log.WithField("database", config.Get().DB.DB).Debug("Added event values")
	for _, grt := range model2.AllGrantTypes {
		if _, err := tx.Exec(`INSERT IGNORE INTO Grants (grant_type) VALUES(?)`, grt); err != nil {
			return err
		}
	}
	log.WithField("database", config.Get().DB.DB).Debug("Added grant_type values")
	return nil
}

func createTables(tx *sqlx.Tx) error {
	if _, err := tx.Exec(`USE ` + config.Get().DB.DB); err != nil {
		return err
	}
	for _, cmd := range dbdefinition.DDL {
		cmd = strings.TrimSpace(cmd)
		if cmd != "" && !strings.HasPrefix(cmd, "--") {
			log.Trace(cmd)
			if _, err := tx.Exec(cmd); err != nil {
				return err
			}
		}
	}
	log.WithField("database", config.Get().DB.DB).Debug("Created tables")
	return nil
}

func createDB(tx *sqlx.Tx) error {
	if _, err := tx.Exec(`DROP DATABASE IF EXISTS ` + config.Get().DB.DB); err != nil {
		return err
	}
	log.WithField("database", config.Get().DB.DB).Debug("Dropped database")
	if _, err := tx.Exec(`CREATE DATABASE ` + config.Get().DB.DB); err != nil {
		return err
	}
	log.WithField("database", config.Get().DB.DB).Debug("Created database")
	return nil
}

func createUser(tx *sqlx.Tx) error {
	log.WithField("user", config.Get().DB.User).Debug("Creating user")
	if _, err := tx.Exec(`CREATE USER IF NOT EXISTS '` + config.Get().DB.User + `' IDENTIFIED BY '` + config.Get().DB.Password + `'`); err != nil {
		return err
	}
	log.WithField("user", config.Get().DB.User).Debug("Created user")
	if _, err := tx.Exec(`GRANT INSERT, UPDATE, DELETE, SELECT ON ` + config.Get().DB.DB + `.* TO '` + config.Get().DB.User + `'`); err != nil {
		return err
	}
	if _, err := tx.Exec(`FLUSH PRIVILEGES `); err != nil {
		return err
	}
	log.WithField("user", config.Get().DB.User).WithField("database", config.Get().DB.DB).Debug("Granted privileges")
	return nil
}

func checkDB(db *cluster.Cluster) error {
	log.WithField("database", config.Get().DB.DB).Debug("Check if database already exists")
	var rows *sql.Rows
	if err := db.Transact(func(tx *sqlx.Tx) error {
		var err error
		rows, err = tx.Query(`SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME=?`, config.Get().DB.DB)
		return err
	}); err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		if !prompter.YesNo("The database already exists. If we continue all data will be deleted. Do you want to continue?", false) {
			rows.Close()
			os.Exit(1) // skipcq CRT-D0011
		}
	}
	return nil
}
