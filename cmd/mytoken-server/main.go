package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/server/config"

	"github.com/zachmann/mytoken/internal/db"
	configurationEndpoint "github.com/zachmann/mytoken/internal/endpoints/configuration"
	"github.com/zachmann/mytoken/internal/jws"
	"github.com/zachmann/mytoken/internal/oidc/authcode"
	server2 "github.com/zachmann/mytoken/internal/server"
	"github.com/zachmann/mytoken/internal/utils/geoip"
	loggerUtils "github.com/zachmann/mytoken/internal/utils/logger"
	"github.com/zachmann/mytoken/shared/httpClient"
)

func main() {
	handleSignals()
	config.Load()
	loggerUtils.Init()
	server2.Init()
	configurationEndpoint.Init()
	authcode.Init()
	if err := db.Connect(); err != nil {
		log.WithError(err).Fatal()
	}
	jws.LoadKey()
	httpClient.Init(config.Get().IssuerURL)
	geoip.Init()

	server2.Start()
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
