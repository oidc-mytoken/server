package main

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/jws"
	loggerUtils "github.com/zachmann/mytoken/internal/utils/logger"
)

func main() {
	config.LoadForSetup()
	loggerUtils.Init()
	sk, _, err := jws.GenerateKeyPair()
	if err != nil {
		log.Fatal(err)
	}
	str := jws.ExportPrivateKeyAsPemStr(sk)
	filepath := config.Get().Signing.KeyFile
	ioutil.WriteFile(filepath, []byte(str), 0600)
	log.WithField("filepath", filepath).Info("Wrote key to file.")
}
