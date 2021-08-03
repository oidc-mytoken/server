package main

import (
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	loggerUtils "github.com/oidc-mytoken/server/internal/utils/logger"
)

func main() {
	config.Load()
	loggerUtils.Init()
	db.Connect()
	deleteExpiredTransferCodes()
	deleteExpiredAuthInfo()
}

func execSimpleQuery(sql string) {
	if err := db.RunWithinTransaction(nil, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(sql)
		return err
	}); err != nil {
		log.WithError(err).Error()
	}
}

func deleteExpiredTransferCodes() {
	execSimpleQuery(
		`DELETE FROM ProxyTokens WHERE id = ANY(SELECT id FROM TransferCodesAttributes
               WHERE expires_at < CURRENT_TIMESTAMP())`)
}

func deleteExpiredAuthInfo() {
	execSimpleQuery(`DELETE FROM AuthInfo WHERE expires_at < CURRENT_TIMESTAMP()`)
}
