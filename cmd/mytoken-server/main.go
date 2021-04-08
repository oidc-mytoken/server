package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	configurationEndpoint "github.com/oidc-mytoken/server/internal/endpoints/configuration"
	"github.com/oidc-mytoken/server/internal/jws"
	"github.com/oidc-mytoken/server/internal/oidc/authcode"
	"github.com/oidc-mytoken/server/internal/server"
	"github.com/oidc-mytoken/server/internal/utils/geoip"
	loggerUtils "github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/shared/httpClient"
)

func main() {
	handleSignals()
	config.Load()
	loggerUtils.Init()
	server.Init()
	configurationEndpoint.Init()
	authcode.Init()
	db.Connect()
	jws.LoadKey()
	httpClient.Init(config.Get().IssuerURL)
	geoip.Init()

	server.Start()
}

func handleSignals() {
	signals := make(chan os.Signal)
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
