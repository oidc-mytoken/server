package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Songmu/prompter"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/jws"
	loggerUtils "github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/internal/utils/zipdownload"
	"github.com/oidc-mytoken/server/shared/utils/fileutil"
)

var genSigningKeyComm commandGenSigningKey
var installComm struct {
	GeoIP commandInstallGeoIPDB `command:"geoip-db" description:"Installs the ip geolocation database."`
}

func main() {
	config.LoadForSetup()
	loggerUtils.Init()

	parser := flags.NewNamedParser("mytoken", flags.HelpFlag|flags.PassDoubleDash)
	if _, err := parser.AddCommand("signing-key", "Generates a new signing key", "Generates a new signing key according to the properties specified in the config file and stores it.", &genSigningKeyComm); err != nil {
		log.WithError(err).Fatal()
		os.Exit(1)
	}
	if _, err := parser.AddCommand("install", "Installs needed dependencies", "", &installComm); err != nil {
		log.WithError(err).Fatal()
		os.Exit(1)
	}
	if _, err := parser.Parse(); err != nil {
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
