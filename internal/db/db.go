package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db/cluster"
)

var db *cluster.Cluster

// Connect connects to the database using the mytoken config
func Connect() {
	ConnectConfig(config.Get().DB)
}

// ConnectConfig connects to the database using the passed config
func ConnectConfig(conf config.DBConf) {
	if db != nil {
		log.Debug("Closing existing db connections")
		db.Close()
	}
	db = cluster.NewFromConfig(conf)
	if err := db.Transact(
		log.StandardLogger(), func(tx *sqlx.Tx) error {
			var err error
			if conf.EnableScheduledCleanup {
				_, err = tx.Exec(`CALL cleanup_schedule_enable()`)
				log.Debug("Enabled scheduled db cleanup")
			} else {
				_, err = tx.Exec(`CALL cleanup_schedule_disable()`)
				log.Debug("Disabled scheduled db cleanup")
			}
			return err
		},
	); err != nil {
		log.WithError(err).Error()
	}

}

// NullString extends the sql.NullString
type NullString struct {
	sql.NullString
}

// MarshalJSON implements the json.Marshaler interface
func (s NullString) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(s.String)
	return data, errors.WithStack(err)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (s *NullString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return errors.WithStack(err)
	}
	*s = NewNullString(str)
	return nil
}

// NewNullString creates a new NullString from the given string
func NewNullString(s string) NullString {
	if s == "" {
		return NullString{}
	}
	return NullString{
		sql.NullString{
			String: s,
			Valid:  true,
		},
	}
}

// NewNullTime creates a new sql.NullTime from the given time.Time
func NewNullTime(t time.Time) sql.NullTime {
	if t.Equal(time.Time{}) {
		return sql.NullTime{}
	}
	return sql.NullTime{
		Time:  t,
		Valid: true,
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
	}
	return []byte{0}, nil
}

// Scan implements the sql.Scanner interface,
// and turns the bitfield incoming from MySQL into a BitBool
func (b *BitBool) Scan(src interface{}) error {
	if src == nil {
		*b = false
		return nil
	}
	v, ok := src.([]byte)
	if !ok {
		return errors.New("bad []byte type assertion")
	}
	*b = v[0] == 1
	return nil
}

// Transact does a database transaction for the passed function
func Transact(rlog log.Ext1FieldLogger, fn func(*sqlx.Tx) error) error {
	return db.Transact(rlog, fn)
}

// RunWithinTransaction runs the passed function using the passed transaction; if nil is passed as tx a new transaction
// is created. This is basically a wrapper function, that works with a possible nil-tx
func RunWithinTransaction(rlog log.Ext1FieldLogger, tx *sqlx.Tx, fn func(*sqlx.Tx) error) error {
	return db.RunWithinTransaction(rlog, tx, fn)
}

// ParseError parses the passed error for a sql.ErrNoRows
func ParseError(e error) (found bool, err error) {
	if e == nil {
		found = true
		return
	}
	if !errors.Is(e, sql.ErrNoRows) {
		err = e
	}
	return
}
