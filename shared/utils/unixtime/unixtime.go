package unixtime

import (
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/oidc-mytoken/server/shared/utils"
)

type UnixTime int64

func (t UnixTime) Time() time.Time {
	return time.Unix(int64(t), 0)
}
func New(t time.Time) UnixTime {
	return UnixTime(t.Unix())
}
func Now() UnixTime {
	return New(time.Now())
}

// InSeconds returns the UnixTime for the current time + the number of passed seconds
func InSeconds(s int64) UnixTime {
	return New(utils.GetTimeIn(s))
}

// Value implements the driver.Valuer interface
func (t UnixTime) Value() (driver.Value, error) {
	return sql.NullTime{
		Time:  t.Time(),
		Valid: true,
	}.Value()
}

// Scan implements the sql.Scanner interface
func (t *UnixTime) Scan(src interface{}) error {
	var tmp sql.NullTime
	if err := tmp.Scan(src); err != nil {
		return err
	}
	*t = New(tmp.Time)
	return nil
}
