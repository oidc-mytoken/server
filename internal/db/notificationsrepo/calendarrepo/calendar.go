package calendarrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/endpoints/notification/calendar/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

// CalendarInfo is a type holding the information stored in the database related to a calendar
type CalendarInfo struct {
	ID          string        `db:"id" json:"id"`
	ICS         string        `db:"ics" json:"-"`
	Description db.NullString `db:"description" json:"description"`
}

// Insert inserts a calendar for the given user (given by the mytoken) into the database
func Insert(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, info CalendarInfo) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Calendar_Insert(?,?,?,?)`, mtID, info.ID, info.Description, info.ICS)
			return errors.WithStack(err)
		},
	)
}

// Delete deletes a calendar for the given user (given by the mytoken) from the database
func Delete(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myid mtid.MTID, calendarID string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Calendar_Delete(?,?)`, myid, calendarID)
			return errors.WithStack(err)
		},
	)
}

// UpdateICS updates a calendar entry in the database
func UpdateICS(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, calendarID, ics string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Calendar_UpdateICS(?,?,?)`, mtID, calendarID, ics)
			return errors.WithStack(err)
		},
	)
}

// UpdateDescription updates the description of a calendar; requires a mytoken id for ownership check
func UpdateDescription(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, calendarID string, description string,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Calendar_UpdateDescription(?,?,?)`, mtID, calendarID, description)
			return errors.WithStack(err)
		},
	)
}

func LinkTags(rlog log.Ext1FieldLogger, tx *sqlx.Tx, calendarID string, tags []string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Calendar_ClearTags(?)`, calendarID)
			if err != nil {
				return errors.WithStack(err)
			}
			for _, tag := range tags {
				_, err = tx.Exec(`CALL Calendar_LinkTag(?,?)`, calendarID, tag)
				if err != nil {
					return errors.WithStack(err)
				}
			}
			return nil
		},
	)
}

// AddTag adds a tag to a calendar
func AddTag(rlog log.Ext1FieldLogger, tx *sqlx.Tx, calendarID, tag string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Calendar_LinkTag(?,?)`, calendarID, tag)
			return errors.WithStack(err)
		},
	)
}

// RemoveTag removes a tag from a calendar
func RemoveTag(rlog log.Ext1FieldLogger, tx *sqlx.Tx, calendarID, tag string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Calendar_UnlinkTag(?,?)`, calendarID, tag)
			return errors.WithStack(err)
		},
	)
}

// UpdateICSInternal updates a calendar entry in the database	 and does not require a mtid.MTID
func UpdateICSInternal(rlog log.Ext1FieldLogger, tx *sqlx.Tx, calendarID, ics string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Calendar_UpdateICSInternal(?,?)`, calendarID, ics)
			return errors.WithStack(err)
		},
	)
}

// GetMTsInCalendar returns a list of mytoken ids that are in a certain calendar (both direct and tag-based)
func GetMTsInCalendar(rlog log.Ext1FieldLogger, tx *sqlx.Tx, calendarID string) (mtids []string, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return tx.Select(&mtids, `CALL Calendar_getMTsInCalendar(?)`, calendarID)
		},
	)
	return
}

// GetDirectMTsInCalendar returns a list of mytoken ids that are directly subscribed to a calendar (not via tags)
func GetDirectMTsInCalendar(rlog log.Ext1FieldLogger, tx *sqlx.Tx, calendarID string) (mtids []string, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return tx.Select(&mtids, `CALL Calendar_GetDirectMTsInCalendar(?)`, calendarID)
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

// GetCalendarTags returns the api.TagInfos for a calendar
func GetCalendarTags(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id string) (tags []api.TagInfo, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&tags, `CALL Calendar_GetTags(?)`, id))
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
				info, err := i.toAPICalendarInfo(rlog, tx)
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

// toAPICalendarInfo transforms a CalendarInfo into an api.CalendarInfo
// Note: SubscribedTokens only contains directly subscribed tokens (not tag-based)
// Tag-based subscriptions can be determined client-side by comparing token tags with calendar tags
func (i CalendarInfo) toAPICalendarInfo(rlog log.Ext1FieldLogger, tx *sqlx.Tx) (
	out api.CalendarInfo, err error,
) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			out = api.CalendarInfo{
				NotificationCalendar: api.NotificationCalendar{
					ID:          i.ID,
					ICSPath:     pkg.GetICSPath(i.ID),
					Description: i.Description.String,
				},
			}
			out.Tags, err = GetCalendarTags(rlog, tx, i.ID)
			if err != nil {
				return err
			}
			// Only return directly subscribed tokens, not tag-based subscriptions
			out.SubscribedTokens, err = GetDirectMTsInCalendar(rlog, tx, i.ID)
			return err
		},
	)
	return
}

// ToCalendarInfoResponse transforms a CalendarInfo into an pkg.CalendarInfoResponse
func (i CalendarInfo) ToCalendarInfoResponse(rlog log.Ext1FieldLogger, tx *sqlx.Tx) (
	*pkg.CalendarInfoResponse, error,
) {
	info, err := i.toAPICalendarInfo(rlog, tx)
	if err != nil {
		return nil, err
	}
	return &pkg.CalendarInfoResponse{
		CalendarInfo: info,
	}, nil

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

// RemoveMytokenFromCalendar removes a mytoken from a calendar in the database
func RemoveMytokenFromCalendar(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, calendarID string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Calendar_RemoveMytoken(?, ?)`, mtID, calendarID)
			return errors.WithStack(err)
		},
	)
}

func MTIsForSameUserAsCalendar(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx,
	calendarID string, mtID mtid.MTID,
) (bool, error) {
	var ok bool
	err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var count int
			err := tx.Get(&count, `CALL Calendar_IDForSameUserAsMT(?, ?)`, calendarID, mtID)
			ok = count > 0
			return err
		},
	)
	return ok, err
}
