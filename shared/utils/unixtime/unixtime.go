package unixtime

import (
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/shared/utils"
)

// UnixTime is a type for a Unix Timestamp
type UnixTime int64

// Time returns the UnixTime as time.Time
func (t UnixTime) Time() time.Time {
	if t == 0 {
		return time.Time{}
	}
	return time.Unix(int64(t), 0)
}

// New creates a new UnixTime from a time.Time
func New(t time.Time) UnixTime {
	return UnixTime(t.Unix())
}

// Now returns the current time as UnixTime
func Now() UnixTime {
	return New(time.Now())
}

// InSeconds returns the UnixTime for the current time + the number of passed seconds
func InSeconds(s int64) UnixTime {
	return New(utils.GetTimeIn(s))
}

// Value implements the driver.Valuer interface
func (t UnixTime) Value() (driver.Value, error) {
	v, err := sql.NullTime{
		Time:  t.Time(),
		Valid: true,
	}.Value()
	return v, errors.WithStack(err)
}

// Scan implements the sql.Scanner interface
func (t *UnixTime) Scan(src interface{}) error {
	var tmp sql.NullTime
	if err := errors.WithStack(tmp.Scan(src)); err != nil {
		return err
	}
	*t = New(tmp.Time)
	return nil
}
