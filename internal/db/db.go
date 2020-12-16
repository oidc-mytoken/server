package db

import (
	"database/sql"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/zachmann/mytoken/internal/config"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var dbCon *sqlx.DB

// Connect connects to the database using the mytoken config
func Connect() error {
	dsn := fmt.Sprintf("%s:%s@%s(%s)/%s", config.Get().DB.User, config.Get().DB.Password, "tcp", config.Get().DB.Host, config.Get().DB.DB)
	dbTmp, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return err
	}
	dbTmp.SetConnMaxLifetime(time.Minute * 4)
	dbTmp.SetMaxOpenConns(10)
	dbTmp.SetMaxIdleConns(10)
	dbCon = dbTmp
	return nil
}

// DB returns the database connection
func DB() *sqlx.DB {
	return dbCon
}

// NewNullString creates a new sql.NullString from the given string
func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

// Transact does a database transaction for the passed function
func Transact(fn func(*sqlx.Tx) error) error {
	tx, err := DB().Beginx()
	if err != nil {
		return err
	}
	err = fn(tx)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			log.WithError(err).Error()
		}
		return err
	}
	return tx.Commit()
}
