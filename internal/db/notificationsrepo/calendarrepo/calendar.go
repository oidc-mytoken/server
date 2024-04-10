package calendarrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

// CalendarInfo is a type holding the information stored in the database related to a calendar
type CalendarInfo struct {
	ID      string `db:"id" json:"id"`
	Name    string `db:"name" json:"name"`
	ICSPath string `db:"ics_path" json:"ics_path"`
	ICS     string `db:"ics" json:"-"`
	pkg.OnlyTokenUpdateRes
}

// Insert inserts a calendar for the given user (given by the mytoken) into the database
func Insert(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, info CalendarInfo) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Calendar_Insert(?,?,?,?,?)`, mtID, info.ID, info.Name, info.ICSPath, info.ICS)
			return errors.WithStack(err)
		},
	)
}

// Delete deletes a calendar for the given user (given by the mytoken) from the database
func Delete(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myid mtid.MTID, name string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Calendar_Delete(?,?)`, myid, name)
			return errors.WithStack(err)
		},
	)
}

// Update updates a calendar entry in the database
func Update(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, info CalendarInfo) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Calendar_Update(?,?,?,?)`, mtID, info.ID, info.Name, info.ICS)
			return errors.WithStack(err)
		},
	)
}

// UpdateInternal updates a calendar entry in the database and does not require a mtid.MTID
func UpdateInternal(rlog log.Ext1FieldLogger, tx *sqlx.Tx, info CalendarInfo) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Calendar_UpdateInternal(?,?,?)`, info.ID, info.Name, info.ICS)
			return errors.WithStack(err)
		},
	)
}

// GetMTsInCalendar returns a list of mytoken ids that are in a certain calendar
func GetMTsInCalendar(rlog log.Ext1FieldLogger, tx *sqlx.Tx, calendarID string) (mtids []string, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return tx.Select(&mtids, `CALL Calendar_getMTsInCalendar(?)`, calendarID)
		},
	)
	return
}

// Get returns a calendar entry for a user and name
func Get(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID any, name string) (info CalendarInfo, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&info, `CALL Calendar_Get(?,?)`, mtID, name))
		},
	)
	return
}

// GetByID returns a calendar entry for a certain calendar id
func GetByID(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id string) (info CalendarInfo, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&info, `CALL Calendar_GetByID(?)`, id))
		},
	)
	return
}

func calendarInfosToAPICalendarInfos(rlog log.Ext1FieldLogger, tx *sqlx.Tx, in []CalendarInfo) (
	out []api.CalendarInfo, err error,
) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			for _, i := range in {
				info := api.CalendarInfo{
					NotificationCalendar: api.NotificationCalendar{
						Name:    i.Name,
						ICSPath: i.ICSPath,
					},
				}
				info.SubscribedTokens, err = GetMTsInCalendar(rlog, tx, i.ID)
				if err != nil {
					return err
				}
				out = append(out, info)
			}
			return nil
		},
	)
	return
}

// List returns a list of all calendar entries for a user
func List(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID) (cals []api.CalendarInfo, err error) {
	var infos []CalendarInfo
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if err = errors.WithStack(tx.Select(&infos, `CALL Calendar_List(?)`, mtID)); err != nil {
				return err
			}
			cals, err = calendarInfosToAPICalendarInfos(rlog, tx, infos)
			return err
		},
	)
	return
}

// ListCalendarsForMT returns a list of calendars where the passed token is used in
func ListCalendarsForMT(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID any) (cals []api.CalendarInfo, err error) {
	var infos []CalendarInfo
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if err = errors.WithStack(tx.Select(&infos, `CALL Calendar_ListForMT(?)`, mtID)); err != nil {
				return err
			}
			cals, err = calendarInfosToAPICalendarInfos(rlog, tx, infos)
			return err
		},
	)
	return
}

// AddMytokenToCalendar associates a mytoken with a calendar in the database; you still have to update the ics
func AddMytokenToCalendar(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, calendarID string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Calendar_AddMytoken(?, ?)`, mtID, calendarID)
			return errors.WithStack(err)
		},
	)
}
