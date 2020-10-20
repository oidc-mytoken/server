package restrictions

import "testing"

func fail(t *testing.T, expected, got Restrictions) {
	t.Errorf("Expected '%v', got '%v'", expected, got)
}

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
	a := Restriction{UsagesAT: 200}
	b := Restriction{UsagesAT: 100}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanUsagesATOneEmpty(t *testing.T) {
	a := Restriction{}
	b := Restriction{UsagesAT: 100}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanUsagesOther(t *testing.T) {
	a := Restriction{UsagesOther: 200}
	b := Restriction{UsagesOther: 100}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanUsagesOtherOneEmpty(t *testing.T) {
	a := Restriction{}
	b := Restriction{UsagesOther: 100}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}

//TODO a test with multiple and all
// NotBefore
// 	ExpiresAt
// 	Scope
// 	Audiences
// 	IPs
// 	GeoIPWhite
// 	GeoIPBlack
// 	UsagesAT
// 	UsagesOther
