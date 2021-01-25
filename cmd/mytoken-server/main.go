package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/httpClient"
	"github.com/zachmann/mytoken/internal/server/config"
	"github.com/zachmann/mytoken/internal/server/db"
	configurationEndpoint "github.com/zachmann/mytoken/internal/server/endpoints/configuration"
	"github.com/zachmann/mytoken/internal/server/jws"
	"github.com/zachmann/mytoken/internal/server/oidc/authcode"
	"github.com/zachmann/mytoken/internal/server/server"
	"github.com/zachmann/mytoken/internal/server/utils/geoip"
	loggerUtils "github.com/zachmann/mytoken/internal/server/utils/logger"
)

func main() {
	handleSignals()
	config.Load()
	loggerUtils.Init()
	server.Init()
	configurationEndpoint.Init()
	authcode.Init()
	if err := db.Connect(); err != nil {
		log.WithError(err).Fatal()
	}
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
	if err := db.Connect(); err != nil {
		log.WithError(err).Fatal()
	}
	jws.LoadKey()
	geoip.Init()
}

func reloadLogFiles() {
	log.Debug("Reloading log files")
	loggerUtils.SetOutput()
	loggerUtils.MustUpdateAccessLogger()
}
