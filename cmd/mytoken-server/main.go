package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/oidc-mytoken/utils/httpclient"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/versionrepo"
	configurationEndpoint "github.com/oidc-mytoken/server/internal/endpoints/configuration"
	"github.com/oidc-mytoken/server/internal/endpoints/settings"
	"github.com/oidc-mytoken/server/internal/jws"
	"github.com/oidc-mytoken/server/internal/model/version"
	"github.com/oidc-mytoken/server/internal/oidc/authcode"
	"github.com/oidc-mytoken/server/internal/server"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/geoip"
	loggerUtils "github.com/oidc-mytoken/server/internal/utils/logger"
)

func main() {
	handleSignals()
	config.Load()
	loggerUtils.Init()
	server.Init()
	configurationEndpoint.Init()
	authcode.Init()
	versionrepo.ConnectToVersion()
	jws.LoadKey()
	httpclient.Init(config.Get().IssuerURL, fmt.Sprintf("mytoken-server %s", version.VERSION))
	geoip.Init()
	settings.InitSettings()
	cookies.Init()

	server.Start()
}

func handleSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGUSR1)
	go func() {
		for {
			sig := <-signals
			switch sig {
			case syscall.SIGHUP:
				reload()
			case syscall.SIGUSR1:
				reloadLogFiles()
			}
		}
	}()
}

func reload() {
	log.Info("Reloading config")
	config.Load()
	loggerUtils.SetOutput()
	loggerUtils.MustUpdateAccessLogger()
	db.Connect()
	jws.LoadKey()
	geoip.Init()
}

func reloadLogFiles() {
	log.Debug("Reloading log files")
	loggerUtils.SetOutput()
	loggerUtils.MustUpdateAccessLogger()
}
