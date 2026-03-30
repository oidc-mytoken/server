package gomigrations

import (
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
)

func init() {
	migrations["v0.11.0"] = func(tx *sqlx.Tx) error {
		type calendarData struct {
			CalendarID string `db:"id"`
			Name       string `db:"name"`
		}
		var data []calendarData
		if err := tx.Select(&data, `SELECT id, name FROM Calendars`); err != nil {
			_, err = db.ParseError(err)
			return err
		}
		for _, d := range data {
			_, err := tx.Exec(`CALL Calendar_LinkTag(?,?)`, d.CalendarID, d.Name)
			if err != nil {
				return err
			}

		}
		return nil
	}
}
