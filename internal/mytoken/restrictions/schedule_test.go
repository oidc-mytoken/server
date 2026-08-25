package restrictions

import (
	"testing"
	"time"

	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/unixtime"
	log "github.com/sirupsen/logrus"
)

func TestDayOrdinal(t *testing.T) {
	utc := time.UTC
	if dayOrdinal(time.Date(2026, 1, 1, 0, 0, 0, 0, utc))+1 != dayOrdinal(time.Date(2026, 1, 2, 0, 0, 0, 0, utc)) {
		t.Error("consecutive days must differ by exactly 1")
	}
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	// DST spring forward happens on 2026-03-29 in Europe/Berlin
	if dayOrdinal(time.Date(2026, 3, 28, 12, 0, 0, 0, berlin))+1 != dayOrdinal(
		time.Date(2026, 3, 29, 12, 0, 0, 0, berlin),
	) {
		t.Error("day ordinal must be independent of DST offsets")
	}
	// 1970-01-01 is day 0
	if dayOrdinal(time.Date(1970, 1, 1, 0, 0, 0, 0, utc)) != 0 {
		t.Errorf("expected 1970-01-01 to be day 0, got %d", dayOrdinal(time.Date(1970, 1, 1, 0, 0, 0, 0, utc)))
	}
}

func TestVerifySchedule(t *testing.T) {
	utc := time.UTC
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	// 2026-01-05 is a Monday, 2026-01-10 a Saturday, 2026-01-11 a Sunday
	mon := time.Date(2026, 1, 5, 10, 0, 0, 0, utc)
	tue := time.Date(2026, 1, 6, 10, 0, 0, 0, utc)
	sat := time.Date(2026, 1, 10, 10, 0, 0, 0, utc)

	tests := []struct {
		name  string
		sched *api.Schedule
		now   time.Time
		want  bool
	}{
		{
			"nil schedule always valid",
			nil,
			mon,
			true,
		},
		{
			"weekdays 9-17 in berlin",
			&api.Schedule{
				Timezone: "Europe/Berlin",
				Days: []int{
					1,
					2,
					3,
					4,
					5,
				},
				From: "09:00",
				To:   "17:00",
			},
			time.Date(2026, 1, 7, 10, 0, 0, 0, berlin),
			true,
		},
		{
			"weekday outside window",
			&api.Schedule{
				Timezone: "Europe/Berlin",
				Days: []int{
					1,
					2,
					3,
					4,
					5,
				},
				From: "09:00",
				To:   "17:00",
			},
			time.Date(2026, 1, 7, 18, 0, 0, 0, berlin),
			false,
		},
		{
			"weekend not allowed",
			&api.Schedule{
				Timezone: "Europe/Berlin",
				Days: []int{
					1,
					2,
					3,
					4,
					5,
				},
				From: "09:00",
				To:   "17:00",
			},
			sat,
			false,
		},
		{
			"only mondays",
			&api.Schedule{Days: []int{1}},
			mon,
			true,
		},
		{
			"only mondays wrong day",
			&api.Schedule{Days: []int{1}},
			tue,
			false,
		},
		{
			"every day 0-1am",
			&api.Schedule{
				From: "00:00",
				To:   "01:00",
			},
			time.Date(2026, 1, 7, 0, 30, 0, 0, utc),
			true,
		},
		{
			"every day outside window",
			&api.Schedule{
				From: "00:00",
				To:   "01:00",
			},
			time.Date(2026, 1, 7, 1, 30, 0, 0, utc),
			false,
		},
		{
			"every sunday 0-1am with 7",
			&api.Schedule{
				Days: []int{7},
				From: "00:00",
				To:   "01:00",
			},
			time.Date(2026, 1, 11, 0, 30, 0, 0, utc),
			true,
		},
		{
			"every sunday 0-1am with 0",
			&api.Schedule{
				Days: []int{0},
				From: "00:00",
				To:   "01:00",
			},
			time.Date(2026, 1, 11, 0, 30, 0, 0, utc),
			true,
		},
		{
			"sunday schedule on monday",
			&api.Schedule{
				Days: []int{0},
				From: "00:00",
				To:   "01:00",
			},
			mon,
			false,
		},
		{
			"every other day from anchor",
			&api.Schedule{
				Every:  2,
				Anchor: "2026-01-01",
				From:   "00:00",
				To:     "01:00",
			},
			time.Date(2026, 1, 3, 0, 30, 0, 0, utc),
			true,
		},
		{
			"every other day off day",
			&api.Schedule{
				Every:  2,
				Anchor: "2026-01-01",
				From:   "00:00",
				To:     "01:00",
			},
			time.Date(2026, 1, 2, 0, 30, 0, 0, utc),
			false,
		},
		{
			"days_of_month 1 and 15",
			&api.Schedule{
				DaysOfMonth: []int{
					1,
					15,
				},
			},
			time.Date(2026, 1, 15, 12, 0, 0, 0, utc),
			true,
		},
		{
			"days_of_month wrong day",
			&api.Schedule{
				DaysOfMonth: []int{
					1,
					15,
				},
			},
			time.Date(2026, 1, 16, 12, 0, 0, 0, utc),
			false,
		},
		{
			"last day of month",
			&api.Schedule{DaysOfMonth: []int{-1}},
			time.Date(2026, 1, 31, 12, 0, 0, 0, utc),
			true,
		},
		{
			"not last day of month",
			&api.Schedule{DaysOfMonth: []int{-1}},
			time.Date(2026, 1, 30, 12, 0, 0, 0, utc),
			false,
		},
		{
			"overnight window inside",
			&api.Schedule{
				Days: []int{1},
				From: "22:00",
				To:   "02:00",
			},
			time.Date(2026, 1, 5, 23, 30, 0, 0, utc),
			true,
		},
		{
			"overnight window spill next day",
			&api.Schedule{
				Days: []int{1},
				From: "22:00",
				To:   "02:00",
			},
			time.Date(2026, 1, 6, 1, 30, 0, 0, utc),
			true,
		},
		{
			"overnight window outside",
			&api.Schedule{
				Days: []int{1},
				From: "22:00",
				To:   "02:00",
			},
			time.Date(2026, 1, 6, 3, 30, 0, 0, utc),
			false,
		},
		{
			"full day via from==to",
			&api.Schedule{
				From: "00:00",
				To:   "00:00",
			},
			time.Date(2026, 1, 7, 15, 0, 0, 0, utc),
			true,
		},
		{
			"invalid timezone",
			&api.Schedule{Timezone: "Not/AZone"},
			time.Date(2026, 1, 7, 15, 0, 0, 0, utc),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := verifySchedule(tt.sched, tt.now); got != tt.want {
					t.Errorf("verifySchedule(%+v, %v) = %v, want %v", tt.sched, tt.now, got, tt.want)
				}
			},
		)
	}
}

func TestScheduleSubset(t *testing.T) {
	tests := []struct {
		name string
		a    *api.Schedule
		b    *api.Schedule
		want bool
	}{
		{
			"equal",
			sched(
				"UTC", "09:00", "17:00", 0, "", []int{
					1,
					2,
					3,
					4,
					5,
				}, nil,
			),
			sched(
				"UTC", "09:00", "17:00", 0, "", []int{
					1,
					2,
					3,
					4,
					5,
				}, nil,
			),
			true,
		},
		{
			"subset weekday",
			sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil),
			sched(
				"UTC", "09:00", "17:00", 0, "", []int{
					1,
					2,
					3,
					4,
					5,
				}, nil,
			),
			true,
		},
		{
			"superset weekday",
			sched(
				"UTC", "09:00", "17:00", 0, "", []int{
					1,
					2,
					3,
					4,
					5,
				}, nil,
			),
			sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil),
			false,
		},
		{
			"wider window",
			sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil),
			sched("UTC", "08:00", "18:00", 0, "", []int{1}, nil),
			true,
		},
		{
			"narrower window",
			sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil),
			sched("UTC", "09:00", "16:00", 0, "", []int{1}, nil),
			false,
		},
		{
			"different timezone",
			sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil),
			sched("Europe/Berlin", "09:00", "17:00", 0, "", []int{1}, nil),
			false,
		},
		{
			"same timezone different alias",
			sched("", "09:00", "17:00", 0, "", []int{1}, nil),
			sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil),
			true,
		},
		{
			"wrap in wrap same days",
			sched("UTC", "22:00", "02:00", 0, "", []int{1}, nil),
			sched(
				"UTC", "22:00", "02:00", 0, "", []int{
					1,
					2,
				}, nil,
			),
			true,
		},
		{
			"wrap in wrap superset days",
			sched(
				"UTC", "22:00", "02:00", 0, "", []int{
					1,
					2,
				}, nil,
			),
			sched("UTC", "22:00", "02:00", 0, "", []int{1}, nil),
			false,
		},
		{
			"wrap subset wrap by window",
			sched(
				"UTC", "22:30", "01:30", 0, "", []int{
					1,
					2,
				}, nil,
			),
			sched(
				"UTC", "22:00", "02:00", 0, "", []int{
					1,
					2,
				}, nil,
			),
			true,
		},
		{
			"wrap not subset wrap by window",
			sched(
				"UTC", "22:00", "02:30", 0, "", []int{
					1,
					2,
				}, nil,
			),
			sched(
				"UTC", "22:00", "02:00", 0, "", []int{
					1,
					2,
				}, nil,
			),
			false,
		},
		{
			"nonwrap in wrap spill",
			sched("UTC", "00:30", "01:00", 0, "", []int{2}, nil),
			sched("UTC", "22:00", "02:00", 0, "", []int{1}, nil),
			true,
		},
		{
			"nonwrap in wrap spill wrong predecessor",
			sched("UTC", "00:30", "01:00", 0, "", []int{3}, nil),
			sched("UTC", "22:00", "02:00", 0, "", []int{1}, nil),
			false,
		},
		{
			"nonwrap in wrap late part",
			sched("UTC", "23:00", "23:30", 0, "", []int{1}, nil),
			sched("UTC", "22:00", "02:00", 0, "", []int{1}, nil),
			true,
		},
		{
			"nonwrap in wrap straddling",
			sched("UTC", "01:00", "23:00", 0, "", []int{2}, nil),
			sched(
				"UTC", "22:00", "02:00", 0, "", []int{
					1,
					2,
				}, nil,
			),
			false,
		},
		{
			"wrap in nonwrap",
			sched("UTC", "22:00", "02:00", 0, "", []int{1}, nil),
			sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil),
			false,
		},
		{
			"anything in full day",
			sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil),
			sched(
				"UTC", "", "", 0, "", []int{
					1,
					2,
				}, nil,
			),
			true,
		},
		{
			"full day not in window",
			sched("UTC", "", "", 0, "", []int{1}, nil),
			sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil),
			false,
		},
		{
			"full day in full day",
			sched("UTC", "", "", 0, "", []int{1}, nil),
			sched(
				"UTC", "", "", 0, "", []int{
					1,
					2,
				}, nil,
			),
			true,
		},
		{
			"full day every day in full day every day",
			sched("UTC", "", "", 0, "", nil, nil),
			sched("UTC", "", "", 0, "", nil, nil),
			true,
		},
		{
			"every same parity",
			sched("UTC", "", "", 2, "2026-01-01", nil, nil),
			sched("UTC", "", "", 2, "2026-01-01", nil, nil),
			true,
		},
		{
			"every wrong parity",
			sched("UTC", "", "", 2, "2026-01-01", nil, nil),
			sched("UTC", "", "", 2, "2026-01-02", nil, nil),
			false,
		},
		{
			"every 4 in every 2",
			sched("UTC", "", "", 4, "2026-01-01", nil, nil),
			sched("UTC", "", "", 2, "2026-01-01", nil, nil),
			true,
		},
		{
			"every 2 not in every 4",
			sched("UTC", "", "", 2, "2026-01-01", nil, nil),
			sched("UTC", "", "", 4, "2026-01-01", nil, nil),
			false,
		},
		{
			"every in every day",
			sched("UTC", "", "", 2, "2026-01-01", nil, nil),
			sched("UTC", "", "", 0, "", nil, nil),
			true,
		},
		{
			"every day not in every",
			sched("UTC", "", "", 0, "", nil, nil),
			sched("UTC", "", "", 2, "2026-01-01", nil, nil),
			false,
		},
		{
			"dom subset",
			sched("UTC", "09:00", "17:00", 0, "", nil, []int{1}),
			sched(
				"UTC", "09:00", "17:00", 0, "", nil, []int{
					1,
					15,
				},
			),
			true,
		},
		{
			"dom superset",
			sched(
				"UTC", "09:00", "17:00", 0, "", nil, []int{
					1,
					15,
				},
			),
			sched("UTC", "09:00", "17:00", 0, "", nil, []int{1}),
			false,
		},
		{
			"dom negative subset",
			sched("UTC", "09:00", "17:00", 0, "", nil, []int{-1}),
			sched(
				"UTC", "09:00", "17:00", 0, "", nil, []int{
					-1,
					-2,
				},
			),
			true,
		},
		{
			"dom negative superset",
			sched(
				"UTC", "09:00", "17:00", 0, "", nil, []int{
					-1,
					-2,
				},
			),
			sched("UTC", "09:00", "17:00", 0, "", nil, []int{-1}),
			false,
		},
		{
			"dom 1 in all days",
			sched("UTC", "09:00", "17:00", 0, "", nil, []int{1}),
			sched("UTC", "09:00", "17:00", 0, "", nil, nil),
			true,
		},
		{
			"all days not in dom 1",
			sched("UTC", "09:00", "17:00", 0, "", nil, nil),
			sched("UTC", "09:00", "17:00", 0, "", nil, []int{1}),
			false,
		},
		{
			"dom in spill branch conservative",
			sched("UTC", "00:30", "01:00", 0, "", nil, []int{15}),
			sched("UTC", "22:00", "02:00", 0, "", []int{1}, nil),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := scheduleSubset(tt.a, tt.b); got != tt.want {
					t.Errorf("scheduleSubset(%+v, %+v) = %v, want %v", tt.a, tt.b, got, tt.want)
				}
			},
		)
	}
}

func sched(tz, from, to string, every int, anchor string, days, dom []int) *api.Schedule {
	return &api.Schedule{
		Timezone:    tz,
		From:        from,
		To:          to,
		Every:       every,
		Anchor:      anchor,
		Days:        days,
		DaysOfMonth: dom,
	}
}

func TestTightenSchedule(t *testing.T) {
	old := Restrictions{
		{Restriction: api.Restriction{Schedule: sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil)}},
	}
	tighter := Restrictions{
		{Restriction: api.Restriction{Schedule: sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil)}},
	}
	res, ok := Tighten(log.StandardLogger(), old, tighter)
	if !ok || len(res) != 1 {
		t.Fatalf("expected ok with one result, got ok=%v len=%d", ok, len(res))
	}
	if res[0].Schedule == nil || len(res[0].Schedule.Days) != 1 || res[0].Schedule.Days[0] != 1 {
		t.Errorf("unexpected result schedule: %+v", res[0].Schedule)
	}

	looser := Restrictions{
		{
			Restriction: api.Restriction{
				Schedule: sched(
					"UTC", "09:00", "17:00", 0, "", []int{
						1,
						2,
						3,
						4,
						5,
					}, nil,
				),
			},
		},
	}
	res2, ok2 := Tighten(log.StandardLogger(), old, looser)
	if ok2 {
		t.Error("expected ok=false when wanted is not tighter")
	}
	if len(res2) != 1 || res2[0].Schedule == nil || len(res2[0].Schedule.Days) != 1 {
		t.Errorf("expected fallback to old restrictions, got %+v", res2)
	}
}

func TestNormalizeSchedule(t *testing.T) {
	s := &api.Schedule{
		Timezone: "",
		Days: []int{
			7,
			0,
			3,
			3,
		},
		DaysOfMonth: []int{
			15,
			15,
			-1,
		},
		Every: 1,
	}
	normalizeSchedule(s)
	if s.Timezone != "UTC" {
		t.Errorf("expected timezone UTC, got %s", s.Timezone)
	}
	if len(s.Days) != 2 || s.Days[0] != 3 || s.Days[1] != 7 {
		t.Errorf("expected days [3 7], got %v", s.Days)
	}
	if len(s.DaysOfMonth) != 2 || s.DaysOfMonth[0] != -1 || s.DaysOfMonth[1] != 15 {
		t.Errorf("expected days_of_month [-1 15], got %v", s.DaysOfMonth)
	}
	if s.Every != 0 {
		t.Errorf("expected every 0, got %d", s.Every)
	}
}

func TestResolveDefaultAnchors(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	creation := time.Date(2026, 1, 15, 10, 0, 0, 0, berlin)
	rs := Restrictions{
		&Restriction{
			Restriction: api.Restriction{
				Schedule: &api.Schedule{
					Timezone: "Europe/Berlin",
					Every:    2,
				},
			},
		},
		&Restriction{
			Restriction: api.Restriction{
				Schedule: &api.Schedule{
					Timezone: "UTC",
					Every:    2,
					Anchor:   "2025-12-01",
				},
			},
		},
		&Restriction{
			Restriction: api.Restriction{
				Schedule: &api.Schedule{
					Timezone: "UTC",
					Every:    0,
				},
			},
		},
		&Restriction{},
	}
	rs.ResolveDefaultAnchors(unixtime.New(creation))
	if rs[0].Schedule.Anchor != "2026-01-15" {
		t.Errorf("expected anchor 2026-01-15, got %s", rs[0].Schedule.Anchor)
	}
	if rs[1].Schedule.Anchor != "2025-12-01" {
		t.Errorf("expected existing anchor to be kept, got %s", rs[1].Schedule.Anchor)
	}
	if rs[2].Schedule.Anchor != "" {
		t.Errorf("expected no anchor for every=0 schedule, got %s", rs[2].Schedule.Anchor)
	}
}

func TestRestrictionScheduleTightening(t *testing.T) {
	mk := func(s *api.Schedule) *Restriction {
		return &Restriction{Restriction: api.Restriction{Schedule: s}}
	}
	// wanted {Mon 9-17} is tighter than old {Mon-Fri 9-17}
	if !mk(sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil)).isTighterThan(
		mk(
			sched(
				"UTC", "09:00", "17:00", 0, "", []int{
					1,
					2,
					3,
					4,
					5,
				}, nil,
			),
		),
	) {
		t.Error("expected tighter")
	}
	// wanted {Mon-Fri 9-17} is not tighter than old {Mon 9-17}
	if mk(
		sched(
			"UTC", "09:00", "17:00", 0, "", []int{
				1,
				2,
				3,
				4,
				5,
			}, nil,
		),
	).isTighterThan(
		mk(sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil)),
	) {
		t.Error("expected not tighter")
	}
	// wanted adds a schedule to a restriction without one
	if !mk(sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil)).isTighterThan(&Restriction{}) {
		t.Error("expected adding a schedule to be tighter")
	}
	// wanted drops the schedule of the old restriction
	if (&Restriction{}).isTighterThan(mk(sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil))) {
		t.Error("expected dropping a schedule to be not tighter")
	}
	// schedule tighter but scope looser still not tighter
	wanted := mk(sched("UTC", "09:00", "17:00", 0, "", []int{1}, nil))
	wanted.Scope = "a b"
	old := mk(
		sched(
			"UTC", "09:00", "17:00", 0, "", []int{
				1,
				2,
			}, nil,
		),
	)
	old.Scope = "a"
	if wanted.isTighterThan(old) {
		t.Error("expected not tighter because of scope")
	}
}
