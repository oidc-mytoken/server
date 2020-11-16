package main

import (
	"log"

	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/db"
)

func main() {
	config.Load()
	if err := db.Connect(); err != nil {
		log.Fatal(err)
	}
	deleteExpiredPollingCodes()
	deleteExpiredAuthInfo()
}

func deleteExpiredPollingCodes() {
	if _, err := db.DB().Exec(`DELETE FROM PollingCodes WHERE expires_at < CURRENT_TIMESTAMP()`); err != nil {
		log.Printf("%s", err)
	}
}

func deleteExpiredAuthInfo() {
	if _, err := db.DB().Exec(`DELETE FROM AuthInfo WHERE expires_at < CURRENT_TIMESTAMP()`); err != nil {
		log.Printf("%s", err)
	}
}
