package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/Songmu/prompter"
	log "github.com/sirupsen/logrus"
	"github.com/zachmann/cli/v2"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/jws"
	"github.com/oidc-mytoken/server/internal/model/version"
	loggerUtils "github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/internal/utils/zipdownload"
	"github.com/oidc-mytoken/server/shared/utils/fileutil"
)

var app = &cli.App{
	Name:     "mytoken-setup",
	Usage:    "Command line client for easily setting up a mytoken server",
	Version:  version.VERSION(),
	Compiled: time.Time{},
	Authors: []*cli.Author{{
		Name:  "Gabriel Zachmann",
		Email: "gabriel.zachmann@kit.edu",
	}},
	Copyright:              "Karlsruhe Institute of Technology 2020-2021",
	UseShortOptionHandling: true,
	Commands: cli.Commands{
		&cli.Command{
			Name:        "signing-key",
			Aliases:     []string{"key"},
			Usage:       "Generates a new signing key",
			Description: "Generates a new signing key according to the properties specified in the config file and stores it.",
			Action:      createSigningKey,
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
	err = ioutil.WriteFile(config.Get().GeoIPDBFile, archive["IP2LOCATION-LITE-DB1.IPV6.BIN"], 0644)
	if err == nil {
		log.WithField("file", config.Get().GeoIPDBFile).Debug("Installed geo ip database")
		fmt.Printf("Installed geo ip database file to '%s'.\n", config.Get().GeoIPDBFile)
	}
	return err
}

func createSigningKey(_ *cli.Context) error {
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
