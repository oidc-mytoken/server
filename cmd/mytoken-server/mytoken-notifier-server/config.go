package main

import (
	"github.com/oidc-mytoken/utils/utils/fileutil"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/oidc-mytoken/server/internal/config"
)

var possibleConfigLocations = []string{
	"config",
	"/etc/mytoken",
}

type conff struct {
	Email config.MailNotificationConf `yaml:"email"`
}

var conf conff

func loadConfig() {
	data, _ := fileutil.MustReadConfigFile("notifier-config.yaml", possibleConfigLocations)
	err := yaml.Unmarshal(data, &conf)
	if err != nil {
		log.WithError(err).Fatal()
		return
	}
}
