package restrictions

import (
	"testing"

	"github.com/zachmann/mytoken/internal/model"
)

func testIsTighter(t *testing.T, a, b Restriction, expected bool) {
	tighter := a.IsTighterThan(b)
	if tighter != expected {
		if expected {
			t.Errorf("Actually '%+v' is tighter than '%+v'", a, b)
		} else {
			t.Errorf("Actually '%+v' is not tighter than '%+v'", a, b)
		}
	}
}

func TestIsTighterThanBothEmpty(t *testing.T) {
	a := Restriction{}
	b := Restriction{}
	testIsTighter(t, a, b, true)
}
func TestIsTighterThanOneEmpty(t *testing.T) {
	a := Restriction{}
	b := Restriction{Scope: "some"}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanNotBefore(t *testing.T) {
	a := Restriction{NotBefore: 50}
	b := Restriction{NotBefore: 100}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanNotBeforeOneEmpty(t *testing.T) {
	a := Restriction{}
	b := Restriction{NotBefore: 100}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanExpiresAt(t *testing.T) {
	a := Restriction{ExpiresAt: 200}
	b := Restriction{ExpiresAt: 100}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanExpiresAtOneEmpty(t *testing.T) {
	a := Restriction{}
	b := Restriction{ExpiresAt: 100}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanScope(t *testing.T) {
	a := Restriction{Scope: "some scopes"}
	b := Restriction{Scope: "some"}
	c := Restriction{Scope: "some other"}
	d := Restriction{Scope: "completely different"}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
	testIsTighter(t, a, c, false)
	testIsTighter(t, b, c, true)
	testIsTighter(t, c, a, false)
	testIsTighter(t, c, b, false)
	testIsTighter(t, a, d, false)
	testIsTighter(t, d, a, false)
}
func TestIsTighterThanScopeOneEmpty(t *testing.T) {
	a := Restriction{}
	b := Restriction{Scope: "some scopes"}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanIP(t *testing.T) {
	a := Restriction{IPs: []string{"192.168.0.12"}}
	b := Restriction{IPs: []string{"192.168.0.12", "192.168.0.14"}}
	c := Restriction{IPs: []string{"192.168.0.0/24"}}
	d := Restriction{IPs: []string{"192.168.1.2", "192.168.0.12"}}
	e := Restriction{IPs: []string{"192.168.0.0/24", "192.168.1.2"}}
	testIsTighter(t, a, b, true)
	testIsTighter(t, b, a, false)
	testIsTighter(t, a, c, true)
	testIsTighter(t, b, c, true)
	testIsTighter(t, c, a, false)
	testIsTighter(t, c, b, false)
	testIsTighter(t, a, d, true)
	testIsTighter(t, d, a, false)
	testIsTighter(t, a, e, true)
	testIsTighter(t, e, a, false)
	testIsTighter(t, d, e, true)
	testIsTighter(t, e, d, false)
}
func TestIsTighterThanIPOneEmpty(t *testing.T) {
	a := Restriction{}
	b := Restriction{IPs: []string{"192.168.0.12", "192.168.0.14"}}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanGeoIPWhite(t *testing.T) {
	a := Restriction{GeoIPWhite: []string{"Germany", "USA"}}
	b := Restriction{GeoIPWhite: []string{"Germany"}}
	c := Restriction{GeoIPWhite: []string{"France", "Germany"}}
	d := Restriction{GeoIPWhite: []string{"Japan", "China"}}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
	testIsTighter(t, a, c, false)
	testIsTighter(t, b, c, true)
	testIsTighter(t, c, a, false)
	testIsTighter(t, c, b, false)
	testIsTighter(t, a, d, false)
	testIsTighter(t, d, a, false)
}
func TestIsTighterThanGeoIPWhiteOneEmpty(t *testing.T) {
	a := Restriction{}
	b := Restriction{GeoIPWhite: []string{"Germany", "USA"}}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanGeoIPBlack(t *testing.T) {
	a := Restriction{GeoIPBlack: []string{"Germany", "USA"}}
	b := Restriction{GeoIPBlack: []string{"Germany"}}
	c := Restriction{GeoIPBlack: []string{"France", "Germany"}}
	d := Restriction{GeoIPBlack: []string{"Japan", "China"}}
	testIsTighter(t, a, b, true)
	testIsTighter(t, b, a, false)
	testIsTighter(t, a, c, false)
	testIsTighter(t, b, c, false)
	testIsTighter(t, c, a, false)
	testIsTighter(t, c, b, true)
	testIsTighter(t, a, d, false)
	testIsTighter(t, d, a, false)
}
func TestIsTighterThanGeoIPBlackOneEmpty(t *testing.T) {
	a := Restriction{}
	b := Restriction{GeoIPBlack: []string{"Germany", "USA"}}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanUsagesAT(t *testing.T) {
	a := Restriction{UsagesAT: model.JSONNullInt{Value: 20, Valid: true}}
	b := Restriction{UsagesAT: model.JSONNullInt{Value: 10, Valid: true}}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanUsagesATOneEmpty(t *testing.T) {
	a := Restriction{}
	b := Restriction{UsagesAT: model.JSONNullInt{Value: 10, Valid: true}}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanUsagesOther(t *testing.T) {
	a := Restriction{UsagesOther: model.JSONNullInt{Value: 20, Valid: true}}
	b := Restriction{UsagesOther: model.JSONNullInt{Value: 10, Valid: true}}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanUsagesOtherOneEmpty(t *testing.T) {
	a := Restriction{}
	b := Restriction{UsagesOther: model.JSONNullInt{Value: 20, Valid: true}}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}

func TestIsTighterThanMultiple1(t *testing.T) {
	a := Restriction{}
	b := Restriction{
		Scope:       "a",
		UsagesAT:    model.JSONNullInt{Value: 50, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 100, Valid: true},
	}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}

func TestIsTighterThanMultiple2(t *testing.T) {
	a := Restriction{
		Scope:       "a",
		UsagesAT:    model.JSONNullInt{Value: 20, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 20, Valid: true},
	}
	b := Restriction{
		Scope:       "a b",
		UsagesAT:    model.JSONNullInt{Value: 50, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 100, Valid: true},
	}
	testIsTighter(t, a, b, true)
	testIsTighter(t, b, a, false)
}

func TestIsTighterThanMultiple3(t *testing.T) {
	a := Restriction{
		UsagesAT:    model.JSONNullInt{Value: 100, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 50, Valid: true},
	}
	b := Restriction{
		UsagesAT:    model.JSONNullInt{Value: 50, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 100, Valid: true},
	}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, false)
}

func TestIsTighterThanMultiple4(t *testing.T) {
	a := Restriction{
		Scope:       "a b c",
		UsagesAT:    model.JSONNullInt{Value: 20, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 20, Valid: true},
	}
	b := Restriction{
		Scope:       "a c b d",
		UsagesAT:    model.JSONNullInt{Value: 50, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 100, Valid: true},
	}
	testIsTighter(t, a, b, true)
	testIsTighter(t, b, a, false)
}
func TestIsTighterThanMultipleE(t *testing.T) {
	a := Restriction{
		Scope:       "a b c",
		UsagesAT:    model.JSONNullInt{Value: 20, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 20, Valid: true},
	}
	b := a
	testIsTighter(t, a, b, true)
	testIsTighter(t, b, a, true)
}

func TestIsTighterThanAll1(t *testing.T) {
	a := Restriction{
		NotBefore:   500,
		ExpiresAt:   1000,
		Scope:       "a b c",
		Audiences:   []string{"a", "b", "c"},
		IPs:         []string{"a", "b", "c"},
		GeoIPWhite:  []string{"a", "b", "c"},
		GeoIPBlack:  []string{"a", "b", "c"},
		UsagesAT:    model.JSONNullInt{Value: 20, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 20, Valid: true},
	}
	b := a
	testIsTighter(t, a, b, true)
	testIsTighter(t, b, a, true)
}

func TestIsTighterThanAll2(t *testing.T) {
	a := Restriction{
		NotBefore:   500,
		ExpiresAt:   1000,
		Scope:       "a b c",
		Audiences:   []string{"a", "b", "c"},
		IPs:         []string{"a", "b", "c"},
		GeoIPWhite:  []string{"a", "b", "c"},
		GeoIPBlack:  []string{"a", "b", "c"},
		UsagesAT:    model.JSONNullInt{Value: 20, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 20, Valid: true},
	}
	b := Restriction{
		NotBefore:   700,
		ExpiresAt:   1000,
		Scope:       "a b c",
		Audiences:   []string{"a", "b", "c"},
		IPs:         []string{"a", "b", "c"},
		GeoIPWhite:  []string{"a", "b", "c"},
		GeoIPBlack:  []string{"a", "b", "c"},
		UsagesAT:    model.JSONNullInt{Value: 20, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 20, Valid: true},
	}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}

func TestIsTighterThanAll3(t *testing.T) {
	a := Restriction{
		NotBefore:   500,
		ExpiresAt:   1000,
		Scope:       "a c",
		Audiences:   []string{"a", "b", "c"},
		IPs:         []string{"a", "b", "c"},
		GeoIPWhite:  []string{"a", "b", "c"},
		GeoIPBlack:  []string{"a", "b", "c"},
		UsagesAT:    model.JSONNullInt{Value: 20, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 20, Valid: true},
	}
	b := Restriction{
		NotBefore:   700,
		ExpiresAt:   1000,
		Scope:       "a b c",
		Audiences:   []string{"a", "b", "c"},
		IPs:         []string{"a", "b", "c"},
		GeoIPWhite:  []string{"a", "b", "c"},
		GeoIPBlack:  []string{"a", "b", "c"},
		UsagesAT:    model.JSONNullInt{Value: 20, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 20, Valid: true},
	}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, false)
}

func TestIsTighterThanAll4(t *testing.T) {
	a := Restriction{
		NotBefore:   500,
		ExpiresAt:   1000,
		Scope:       "a b c",
		Audiences:   []string{"a", "b", "c"},
		IPs:         []string{"a", "b", "c"},
		GeoIPWhite:  []string{"a", "b", "c"},
		GeoIPBlack:  []string{"a", "b", "c"},
		UsagesAT:    model.JSONNullInt{Value: 20, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 20, Valid: true},
	}
	b := Restriction{
		NotBefore:   700,
		ExpiresAt:   900,
		Scope:       "b c",
		Audiences:   []string{"a", "c"},
		IPs:         []string{"b", "c"},
		GeoIPWhite:  []string{"a"},
		GeoIPBlack:  []string{"a", "b"},
		UsagesAT:    model.JSONNullInt{Value: 10, Valid: true},
		UsagesOther: model.JSONNullInt{Value: 0, Valid: true},
	}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, false)
}
