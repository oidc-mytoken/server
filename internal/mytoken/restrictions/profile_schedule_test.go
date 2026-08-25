package restrictions

import (
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/profilerepo"
)

func TestParseScheduleProfile(t *testing.T) {
	parser := profilerepo.NewDBProfileParser(log.StandardLogger())
	r, err := parser.ParseRestrictionsTemplate([]byte(`[{"schedule":{"timezone":"Europe/Berlin","days":[1,2,3,4,5],"from":"09:00","to":"17:00"}}]`))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(r) != 1 || r[0].Schedule == nil {
		t.Fatalf("expected schedule in parsed restriction")
	}
	rs := NewRestrictionsFromAPI(r)
	if rs[0].Schedule.Timezone != "Europe/Berlin" {
		t.Errorf("unexpected schedule: %+v", rs[0].Schedule)
	}
	berlin, _ := time.LoadLocation("Europe/Berlin")
	if !verifySchedule(rs[0].Schedule, time.Date(2026, 1, 7, 10, 0, 0, 0, berlin)) {
		t.Error("expected schedule to allow a weekday at 10:00")
	}
	if _, err := parser.ParseRestrictionsTemplate([]byte(`[{"schedule":{"days":[9]}}]`)); err == nil {
		t.Error("expected error for invalid schedule")
	}
}
