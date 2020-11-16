package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/db"
	loggerUtils "github.com/zachmann/mytoken/internal/utils/logger"
)

func main() {
	config.Load()
	loggerUtils.Init()
	if err := db.Connect(); err != nil {
		log.WithError(err).Fatal()
	}
	deleteExpiredPollingCodes()
	deleteExpiredAuthInfo()
}

func deleteExpiredPollingCodes() {
	if _, err := db.DB().Exec(`DELETE FROM PollingCodes WHERE expires_at < CURRENT_TIMESTAMP()`); err != nil {
		log.WithError(err).Error()
	}
}

func deleteExpiredAuthInfo() {
	if _, err := db.DB().Exec(`DELETE FROM AuthInfo WHERE expires_at < CURRENT_TIMESTAMP()`); err != nil {
		log.WithError(err).Error()
	}
}
