package restrictions

import (
	"testing"

	"github.com/zachmann/mytoken/internal/model"
)

// import "testing"

//TODO TestTighten

func checkRestrictions(t *testing.T, exp, a Restrictions) {
	if len(a) != len(exp) {
		t.Errorf("Expected '%+v', but got '%+v'", exp, a)
		return
	}
	for i, ee := range exp {
		aa := a[i]
		if !(ee.IsTighterThan(aa) && aa.IsTighterThan(ee)) {
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
			UsagesAT: model.JSONNullInt{
				Value: 10,
				Valid: true,
			},
		},
	}
	wanted := Restrictions{
		{
			UsagesAT: model.JSONNullInt{
				Value: 10,
				Valid: true,
			},
		},
		{
			UsagesAT: model.JSONNullInt{
				Value: 10,
				Valid: true,
			},
		},
	}
	expected := base
	res := Tighten(base, wanted)
	checkRestrictions(t, expected, res)
}
