package restrictions

import (
	"slices"
	"sort"
	"time"

	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/unixtime"
)

func loadLocation(tz string) (*time.Location, error) {
	if tz == "" || tz == "UTC" {
		return time.UTC, nil
	}
	return time.LoadLocation(tz)
}

func tzName(tz string) string {
	if tz == "" || tz == "UTC" {
		return "UTC"
	}
	return tz
}

func sameTimezone(a, b string) bool {
	return tzName(a) == tzName(b)
}

// normalizeSchedule canonicalizes a schedule in place. It must be called before a schedule
// is stored or used in comparisons.
func normalizeSchedule(s *api.Schedule) {
	if s == nil {
		return
	}
	if s.Timezone == "" {
		s.Timezone = "UTC"
	}
	s.Days = normalizeDays(s.Days)
	s.DaysOfMonth = normalizeDom(s.DaysOfMonth)
	if s.Every <= 1 {
		s.Every = 0
	}
}

// normalizeDays normalizes the weekdays of a schedule: 0 and 7 are both Sunday and are
// collapsed to 7, duplicates are removed and the result is sorted.
func normalizeDays(days []int) []int {
	seen := make(map[int]bool, len(days))
	ret := make([]int, 0, len(days))
	for _, d := range days {
		if d == 0 {
			d = 7
		}
		if d < 1 || d > 7 || seen[d] {
			continue
		}
		seen[d] = true
		ret = append(ret, d)
	}
	sort.Ints(ret)
	return ret
}

// normalizeDom normalizes the days of month of a schedule: duplicates are removed and the
// result is sorted.
func normalizeDom(doms []int) []int {
	seen := make(map[int]bool, len(doms))
	ret := make([]int, 0, len(doms))
	for _, d := range doms {
		if d == 0 || d < -31 || d > 31 || seen[d] {
			continue
		}
		seen[d] = true
		ret = append(ret, d)
	}
	sort.Ints(ret)
	return ret
}

func mod(x, m int64) int64 {
	r := x % m
	if r < 0 {
		r += m
	}
	return r
}

// dayOrdinal returns the number of days since 1970-01-01 for the given date. It is
// independent of timezone offsets, so two consecutive calendar days always differ by
// exactly 1.
func dayOrdinal(t time.Time) int64 {
	y, m, d := t.Date()
	yy := int64(y)
	mm := int64(m)
	if mm <= 2 {
		yy--
		mm += 9
	} else {
		mm -= 3
	}
	era := yy / 400
	yoe := yy - era*400
	doy := (153*mm+2)/5 + int64(d) - 1
	doe := yoe*365 + yoe/4 - yoe/100 + doy
	return era*146097 + doe - 719468
}

func parseTimeOfDay(s string) int {
	t, _ := time.Parse("15:04", s)
	return t.Hour()*60 + t.Minute()
}

// scheduleWindow returns the from and to minutes of a schedule's window. fullDay is true
// when the window covers the whole day (from and to are equal or not set).
func scheduleWindow(s *api.Schedule) (from, to int, fullDay bool) {
	if s.From == "" || s.To == "" {
		return 0, 0, true
	}
	from, to = parseTimeOfDay(s.From), parseTimeOfDay(s.To)
	return from, to, from == to
}

func daysInMonth(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
}

func domMatches(doms []int, local time.Time) bool {
	d := local.Day()
	for _, x := range doms {
		if x > 0 {
			if d == x {
				return true
			}
		} else if d == daysInMonth(local)+x+1 {
			return true
		}
	}
	return false
}

func scheduleDayMatches(s *api.Schedule, local time.Time) bool {
	days := normalizeDays(s.Days)
	wd := int(local.Weekday())
	if wd == 0 {
		wd = 7
	}
	if len(days) > 0 && !slices.Contains(days, wd) {
		return false
	}
	if len(s.DaysOfMonth) > 0 && !domMatches(s.DaysOfMonth, local) {
		return false
	}
	if s.Every > 0 {
		anchorDate, err := time.ParseInLocation("2006-01-02", s.Anchor, local.Location())
		if err != nil {
			return false
		}
		if mod(dayOrdinal(local)-dayOrdinal(anchorDate), int64(s.Every)) != 0 {
			return false
		}
	}
	return true
}

// verifySchedule checks that now is within the periodic time window described by the
// schedule.
func verifySchedule(s *api.Schedule, now time.Time) bool {
	if s == nil {
		return true
	}
	loc, err := loadLocation(s.Timezone)
	if err != nil {
		return false
	}
	local := now.In(loc)
	return scheduleAllows(s, local)
}

// scheduleAllows checks whether the given local time is allowed by the schedule. For a
// wrapping window ([from, to) with to < from) the early-morning part of a day is only
// allowed if the previous day is selected by the schedule.
func scheduleAllows(s *api.Schedule, local time.Time) bool {
	from, to, fullDay := scheduleWindow(s)
	if fullDay {
		return scheduleDayMatches(s, local)
	}
	now := local.Hour()*60 + local.Minute()
	if from < to {
		return scheduleDayMatches(s, local) && now >= from && now < to
	}
	if now >= from {
		return scheduleDayMatches(s, local)
	}
	if now < to {
		prev := local.AddDate(0, 0, -1)
		return scheduleDayMatches(s, prev)
	}
	return false
}

// ResolveDefaultAnchors sets the anchor date of schedules with an every interval to the
// given creation time if no anchor was specified.
func (r *Restrictions) ResolveDefaultAnchors(now unixtime.UnixTime) {
	for _, rr := range *r {
		if rr.Schedule == nil || rr.Schedule.Every <= 0 || rr.Schedule.Anchor != "" {
			continue
		}
		loc, err := loadLocation(rr.Schedule.Timezone)
		if err != nil {
			continue
		}
		rr.Schedule.Anchor = now.Time().In(loc).Format("2006-01-02")
	}
}

// scheduleOrdinal returns the day ordinal of a schedule's anchor in the given location,
// or 0 if the anchor is not set or not parseable.
func scheduleOrdinal(s *api.Schedule, loc *time.Location) int64 {
	if s.Anchor == "" {
		return 0
	}
	t, err := time.ParseInLocation("2006-01-02", s.Anchor, loc)
	if err != nil {
		return 0
	}
	return dayOrdinal(t)
}

// scheduleDaySubset reports whether every day selected by a is also selected by b.
func scheduleDaySubset(a, b *api.Schedule) bool {
	if !dayComponentSubset(a.Days, b.Days) {
		return false
	}
	if !dayComponentSubset(a.DaysOfMonth, b.DaysOfMonth) {
		return false
	}
	if a.Every <= 0 {
		return b.Every <= 0
	}
	if b.Every <= 0 {
		return true
	}
	if a.Every%b.Every != 0 {
		return false
	}
	if a.Anchor == "" || b.Anchor == "" {
		return false
	}
	loc, err := loadLocation(a.Timezone)
	if err != nil {
		return false
	}
	return mod(scheduleOrdinal(a, loc)-scheduleOrdinal(b, loc), int64(b.Every)) == 0
}

// dayComponentSubset reports whether aSel is a subset of bSel, where an empty slice
// selects all values.
func dayComponentSubset(aSel, bSel []int) bool {
	if len(aSel) == 0 {
		return len(bSel) == 0
	}
	if len(bSel) == 0 {
		return true
	}
	for _, x := range aSel {
		if !slices.Contains(bSel, x) {
			return false
		}
	}
	return true
}

// schedulePredecessorCovered reports whether every day selected by a has its predecessor
// selected by b. This is needed to verify that a non-wrapping window of a can be covered
// by the overnight spill of a wrapping window of b.
func schedulePredecessorCovered(a, b *api.Schedule) bool {
	if !weekdayPredecessorCovered(a.Days, b.Days) {
		return false
	}
	if len(b.DaysOfMonth) > 0 {
		return false // conservative: the predecessor may cross a month boundary
	}
	if b.Every <= 0 {
		return true
	}
	if a.Every <= 0 || a.Anchor == "" || b.Anchor == "" || a.Every%b.Every != 0 {
		return false
	}
	loc, err := loadLocation(a.Timezone)
	if err != nil {
		return false
	}
	return mod(scheduleOrdinal(a, loc)-1-scheduleOrdinal(b, loc), int64(b.Every)) == 0
}

func weekdayPredecessorCovered(aDays, bDays []int) bool {
	if len(aDays) == 0 {
		return len(bDays) == 0
	}
	if len(bDays) == 0 {
		return true
	}
	for _, d := range aDays {
		pred := d - 1
		if pred == 0 {
			pred = 7
		}
		if !slices.Contains(bDays, pred) {
			return false
		}
	}
	return true
}

// scheduleSubset reports whether a's allowed set is a subset of b's allowed set.
func scheduleSubset(a, b *api.Schedule) bool {
	if !sameTimezone(a.Timezone, b.Timezone) {
		return false
	}
	aFrom, aTo, aFull := scheduleWindow(a)
	bFrom, bTo, bFull := scheduleWindow(b)
	aWrap, bWrap := aFrom > aTo, bFrom > bTo
	dayOK := scheduleDaySubset(a, b)
	predecOK := schedulePredecessorCovered(a, b)
	switch {
	case aFull:
		return bFull && dayOK
	case bFull:
		return dayOK
	case !aWrap && !bWrap:
		return dayOK && aFrom >= bFrom && aTo <= bTo
	case !aWrap: // -> bWrap == true
		// a's window is either covered by b's late-day part on the same day (needs b to be
		// active on that day) or by b's overnight spill (needs the predecessor day to be
		// active for b).
		return (dayOK && aFrom >= bFrom) || (predecOK && aTo <= bTo)
	case !bWrap: // aWrap == true
		return false // a wrapping window can never be covered by a non-wrapping one
	default:
		return dayOK && aFrom >= bFrom && aTo <= bTo
	}
}
