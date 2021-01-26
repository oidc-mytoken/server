package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/server/config"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var dbCon *sqlx.DB

// Connect connects to the database using the mytoken config
func Connect() error {
	dsn := fmt.Sprintf("%s:%s@%s(%s)/%s", config.Get().DB.User, config.Get().DB.Password, "tcp", config.Get().DB.Host, config.Get().DB.DB)
	return ConnectDSN(dsn)
}

// ConnectDSN connects to a database using the dsn string
func ConnectDSN(dsn string) error {
	dbTmp, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return err
	}
	dbTmp.SetConnMaxLifetime(time.Minute * 4)
	dbTmp.SetMaxOpenConns(10)
	dbTmp.SetMaxIdleConns(10)
	if dbCon != nil {
		if err = dbCon.Close(); err != nil {
			log.WithError(err).Error()
		}
	}
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

// BitBool is an implementation of a bool for the MySQL type BIT(1).
// This type allows you to avoid wasting an entire byte for MySQL's boolean type TINYINT.
type BitBool bool

// Value implements the driver.Valuer interface,
// and turns the BitBool into a bitfield (BIT(1)) for MySQL storage.
func (b BitBool) Value() (driver.Value, error) {
	if b {
		return []byte{1}, nil
	} else {
		return []byte{0}, nil
	}
}

// Scan implements the sql.Scanner interface,
// and turns the bitfield incoming from MySQL into a BitBool
func (b *BitBool) Scan(src interface{}) error {
	v, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("bad []byte type assertion")
	}
	*b = v[0] == 1
	return nil
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

// RunWithinTransaction runs the passed function using the passed transaction; if nil is passed as tx a new transaction is created. This is basically a wrapper function, that works with a possible nil-tx
func RunWithinTransaction(tx *sqlx.Tx, fn func(*sqlx.Tx) error) error {
	if tx == nil {
		return Transact(fn)
	} else {
		return fn(tx)
	}
}
