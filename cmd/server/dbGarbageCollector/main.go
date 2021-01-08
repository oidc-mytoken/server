package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/server/config"
	"github.com/zachmann/mytoken/internal/server/db"
	loggerUtils "github.com/zachmann/mytoken/internal/server/utils/logger"
)

func main() {
	config.Load()
	loggerUtils.Init()
	if err := db.Connect(); err != nil {
		log.WithError(err).Fatal()
	}
	deleteExpiredTransferCodes()
	deleteExpiredAuthInfo()
}

func deleteExpiredTransferCodes() {
	if _, err := db.DB().Exec(`DELETE FROM ProxyTokens WHERE id = ANY(SELECT id FROM TransferCodesAttributes WHERE expires_at < CURRENT_TIMESTAMP())`); err != nil {
		log.WithError(err).Error()
	}
}

func deleteExpiredAuthInfo() {
	if _, err := db.DB().Exec(`DELETE FROM AuthInfo WHERE expires_at < CURRENT_TIMESTAMP()`); err != nil {
		log.WithError(err).Error()
	}
}
