package restrictions

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/oidc-mytoken/server/shared/utils"
)

func checkRestrictions(t *testing.T, exp, a Restrictions) {
	if len(a) != len(exp) {
		t.Errorf("Expected '%+v', but got '%+v'", exp, a)
		return
	}
	for i, ee := range exp {
		aa := a[i]
		if !(ee.isTighterThan(aa) && aa.isTighterThan(ee)) {
			t.Errorf("Expected '%+v', but got '%+v'", exp, a)
			return
		}
	}
}
func TestTighten_RestrictEmpty(t *testing.T) {
	base := Restrictions{}
	wanted := Restrictions{
		{
			Scope:     "a b c",
			ExpiresAt: 100,
		},
	}
	expected := wanted
	res := Tighten(base, wanted)
	checkRestrictions(t, expected, res)
}
func TestTighten_RequestEmpty(t *testing.T) {
	base := Restrictions{
		{
			Scope:     "a b c",
			ExpiresAt: 500,
		},
		{
			Scope:     "a",
			ExpiresAt: 1000,
		},
		{
			Scope:     "d",
			ExpiresAt: 50,
		},
	}
	wanted := Restrictions{}
	expected := base
	res := Tighten(base, wanted)
	checkRestrictions(t, expected, res)
}
func TestTighten_RestrictToOne(t *testing.T) {
	base := Restrictions{
		{
			Scope:     "a b c",
			ExpiresAt: 500,
		},
		{
			Scope:     "a",
			ExpiresAt: 1000,
		},
		{
			Scope:     "d",
			ExpiresAt: 50,
		},
	}
	wanted := Restrictions{
		{
			Scope:     "a b c",
			ExpiresAt: 100,
		},
	}
	expected := wanted
	res := Tighten(base, wanted)
	checkRestrictions(t, expected, res)
}
func TestTighten_RestrictToTwo(t *testing.T) {
	base := Restrictions{
		{
			Scope:     "a b c",
			ExpiresAt: 500,
		},
		{
			Scope:     "a",
			ExpiresAt: 1000,
		},
		{
			Scope:     "d",
			ExpiresAt: 50,
		},
	}
	wanted := Restrictions{
		{
			Scope:     "a b c",
			ExpiresAt: 100,
		},
		{
			Scope:     "d",
			ExpiresAt: 50,
		},
	}
	expected := wanted
	res := Tighten(base, wanted)
	checkRestrictions(t, expected, res)
}
func TestTighten_RestrictConflict(t *testing.T) {
	base := Restrictions{
		{
			Scope:     "a b c",
			ExpiresAt: 500,
		},
		{
			Scope:     "a",
			ExpiresAt: 1000,
		},
		{
			Scope:     "d",
			ExpiresAt: 50,
		},
	}
	wanted := Restrictions{
		{
			Scope:     "a b c d",
			ExpiresAt: 100,
		},
	}
	expected := base
	res := Tighten(base, wanted)
	checkRestrictions(t, expected, res)
}
func TestTighten_RestrictDontCombineTwo(t *testing.T) {
	base := Restrictions{
		{
			Scope:     "a b c",
			ExpiresAt: 500,
		},
		{
			Scope:     "d",
			ExpiresAt: 50,
		},
	}
	wanted := Restrictions{
		{
			Scope:     "a b c d", // This is semantically different from base, because it allows a token with all the scopes combined. One might want to not allow this.
			ExpiresAt: 50,
		},
	}
	expected := base
	res := Tighten(base, wanted)
	checkRestrictions(t, expected, res)
}
func TestTighten_RestrictDontExtendUsages(t *testing.T) {
	base := Restrictions{
		{
			UsagesAT: utils.NewInt64(10),
		},
	}
	wanted := Restrictions{
		{
			UsagesAT: utils.NewInt64(10),
		},
		{
			UsagesAT: utils.NewInt64(10),
		},
	}
	expected := base
	res := Tighten(base, wanted)
	checkRestrictions(t, expected, res)
}

func testIsTighter(t *testing.T, a, b Restriction, expected bool) {
	tighter := a.isTighterThan(b)
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
	a := Restriction{UsagesAT: utils.NewInt64(20)}
	b := Restriction{UsagesAT: utils.NewInt64(10)}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanUsagesATOneEmpty(t *testing.T) {
	a := Restriction{}
	b := Restriction{UsagesAT: utils.NewInt64(10)}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanUsagesOther(t *testing.T) {
	a := Restriction{UsagesOther: utils.NewInt64(20)}
	b := Restriction{UsagesOther: utils.NewInt64(10)}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanUsagesOtherOneEmpty(t *testing.T) {
	a := Restriction{}
	b := Restriction{UsagesOther: utils.NewInt64(20)}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanMultiple1(t *testing.T) {
	a := Restriction{}
	b := Restriction{
		Scope:       "a",
		UsagesAT:    utils.NewInt64(50),
		UsagesOther: utils.NewInt64(100),
	}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, true)
}
func TestIsTighterThanMultiple2(t *testing.T) {
	a := Restriction{
		Scope:       "a",
		UsagesAT:    utils.NewInt64(20),
		UsagesOther: utils.NewInt64(20),
	}
	b := Restriction{
		Scope:       "a b",
		UsagesAT:    utils.NewInt64(50),
		UsagesOther: utils.NewInt64(100),
	}
	testIsTighter(t, a, b, true)
	testIsTighter(t, b, a, false)
}
func TestIsTighterThanMultiple3(t *testing.T) {
	a := Restriction{
		UsagesAT:    utils.NewInt64(100),
		UsagesOther: utils.NewInt64(50),
	}
	b := Restriction{
		UsagesAT:    utils.NewInt64(50),
		UsagesOther: utils.NewInt64(100),
	}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, false)
}
func TestIsTighterThanMultiple4(t *testing.T) {
	a := Restriction{
		Scope:       "a b c",
		UsagesAT:    utils.NewInt64(20),
		UsagesOther: utils.NewInt64(20),
	}
	b := Restriction{
		Scope:       "a c b d",
		UsagesAT:    utils.NewInt64(50),
		UsagesOther: utils.NewInt64(100),
	}
	testIsTighter(t, a, b, true)
	testIsTighter(t, b, a, false)
}
func TestIsTighterThanMultipleE(t *testing.T) {
	a := Restriction{
		Scope:       "a b c",
		UsagesAT:    utils.NewInt64(20),
		UsagesOther: utils.NewInt64(20),
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
		UsagesAT:    utils.NewInt64(20),
		UsagesOther: utils.NewInt64(20),
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
		UsagesAT:    utils.NewInt64(20),
		UsagesOther: utils.NewInt64(20),
	}
	b := Restriction{
		NotBefore:   700,
		ExpiresAt:   1000,
		Scope:       "a b c",
		Audiences:   []string{"a", "b", "c"},
		IPs:         []string{"a", "b", "c"},
		GeoIPWhite:  []string{"a", "b", "c"},
		GeoIPBlack:  []string{"a", "b", "c"},
		UsagesAT:    utils.NewInt64(20),
		UsagesOther: utils.NewInt64(20),
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
		UsagesAT:    utils.NewInt64(20),
		UsagesOther: utils.NewInt64(20),
	}
	b := Restriction{
		NotBefore:   700,
		ExpiresAt:   1000,
		Scope:       "a b c",
		Audiences:   []string{"a", "b", "c"},
		IPs:         []string{"a", "b", "c"},
		GeoIPWhite:  []string{"a", "b", "c"},
		GeoIPBlack:  []string{"a", "b", "c"},
		UsagesAT:    utils.NewInt64(20),
		UsagesOther: utils.NewInt64(20),
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
		UsagesAT:    utils.NewInt64(20),
		UsagesOther: utils.NewInt64(20),
	}
	b := Restriction{
		NotBefore:   700,
		ExpiresAt:   900,
		Scope:       "b c",
		Audiences:   []string{"a", "c"},
		IPs:         []string{"b", "c"},
		GeoIPWhite:  []string{"a"},
		GeoIPBlack:  []string{"a", "b"},
		UsagesAT:    utils.NewInt64(10),
		UsagesOther: utils.NewInt64(0),
	}
	testIsTighter(t, a, b, false)
	testIsTighter(t, b, a, false)
}

func TestRestrictions_GetExpiresEmpty(t *testing.T) {
	r := Restrictions{}
	expires := r.GetExpires()
	var expected int64 = 0
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}
func TestRestrictions_GetExpiresInfinite(t *testing.T) {
	r := Restrictions{
		{ExpiresAt: 0},
	}
	expires := r.GetExpires()
	var expected int64 = 0
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}
func TestRestrictions_GetExpiresOne(t *testing.T) {
	r := Restrictions{
		{ExpiresAt: 100},
	}
	expires := r.GetExpires()
	var expected int64 = 100
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}
func TestRestrictions_GetExpiresMultiple(t *testing.T) {
	r := Restrictions{
		{ExpiresAt: 100},
		{ExpiresAt: 300},
		{ExpiresAt: 200},
	}
	expires := r.GetExpires()
	var expected int64 = 300
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}
func TestRestrictions_GetExpiresMultipleAndInfinite(t *testing.T) {
	r := Restrictions{
		{ExpiresAt: 100},
		{ExpiresAt: 0},
		{ExpiresAt: 300},
		{ExpiresAt: 200},
	}
	expires := r.GetExpires()
	var expected int64 = 0
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}

func TestRestrictions_GetNotBeforeEmpty(t *testing.T) {
	r := Restrictions{}
	expires := r.GetNotBefore()
	var expected int64 = 0
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}
func TestRestrictions_GetNotBeforeInfinite(t *testing.T) {
	r := Restrictions{
		{NotBefore: 0},
	}
	expires := r.GetNotBefore()
	var expected int64 = 0
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}
func TestRestrictions_GetNotBeforeOne(t *testing.T) {
	r := Restrictions{
		{NotBefore: 100},
	}
	expires := r.GetNotBefore()
	var expected int64 = 100
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}
func TestRestrictions_GetNotBeforeMultiple(t *testing.T) {
	r := Restrictions{
		{NotBefore: 100},
		{NotBefore: 300},
		{NotBefore: 200},
	}
	expires := r.GetNotBefore()
	var expected int64 = 100
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}
func TestRestrictions_GetNotBeforeMultipleAndInfinite(t *testing.T) {
	r := Restrictions{
		{NotBefore: 100},
		{NotBefore: 0},
		{NotBefore: 300},
		{NotBefore: 200},
	}
	expires := r.GetNotBefore()
	var expected int64 = 0
	if expected != expires {
		t.Errorf("Expected %d, but got %d", expected, expires)
	}
}

func TestRestriction_Hash(t *testing.T) {
	r := Restriction{
		NotBefore: 1599939600,
		ExpiresAt: 1599948600,
		IPs:       []string{"192.168.0.31"},
		UsagesAT:  utils.NewInt64(11),
	}

	j, err := json.Marshal(r)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n", j)

	hash, err := r.hash()
	if err != nil {
		t.Error(err)
	}
	expected := "BS3WfHbHNUiVU8sJ+F49H9+69HnFtfVDy2m22vBv588nZ0kGblVNxZEcrTN+5NUiRkM7W80N4VpPgwEZBZl+3g=="
	if string(hash) != expected {
		t.Errorf("hash '%s' does not match expected hash '%s'", hash, expected)
	}
}

func TestRestriction_VerifyTimeBased(t *testing.T) {
	now := utils.GetUnixTimeIn(0)
	cases := []struct {
		r   Restriction
		exp bool
	}{
		{
			r:   Restriction{NotBefore: 0, ExpiresAt: 0},
			exp: true,
		},
		{
			r:   Restriction{NotBefore: now - 10, ExpiresAt: 0},
			exp: true,
		},
		{
			r:   Restriction{NotBefore: 0, ExpiresAt: now + 10},
			exp: true,
		},
		{
			r:   Restriction{NotBefore: now + 10, ExpiresAt: 0},
			exp: false,
		},
		{
			r:   Restriction{NotBefore: 0, ExpiresAt: now - 10},
			exp: false,
		},
		{
			r:   Restriction{NotBefore: now + 10, ExpiresAt: now - 10},
			exp: false,
		},
	}
	for _, c := range cases {
		valid := c.r.verifyTimeBased()
		if valid != c.exp {
			t.Errorf("For '%+v' expected time based attributes to verify as '%v' but got '%v'", c.r, c.exp, valid)
		}
	}
}
