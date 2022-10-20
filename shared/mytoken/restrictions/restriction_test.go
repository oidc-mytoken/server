package restrictions

import (
	"testing"

	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/shared/utils"
	"github.com/oidc-mytoken/server/shared/utils/unixtime"
)

func checkRestrictions(t *testing.T, exp, a Restrictions, okExp, ok bool) {
	if okExp != ok {
		t.Errorf("Expected '%+v', but got '%+v' for ok-status", okExp, ok)
	}
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

type tightenTestCase struct {
	name     string
	base     Restrictions
	wanted   Restrictions
	expected Restrictions
	okExp    bool
}

func TestTighten(t *testing.T) {

	tests := []tightenTestCase{
		tightenCase1(),
		tightenCase2(),
		tightenCase3(),
		tightenCase4(),
		tightenCase5(),
		tightenCase6(),
		tightenCase7(),
		tightenCase8(),
		tightenCase9(),
		tightenCase10(),
		tightenCase11(),
	}

	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				res, ok := Tighten(log.StandardLogger(), test.base, test.wanted)
				checkRestrictions(t, test.expected, res, test.okExp, ok)
			},
		)
	}
}

func tightenCase1() tightenTestCase {
	test := tightenTestCase{
		name: "Empty Base",
		base: Restrictions{},
		wanted: Restrictions{
			{
				Restriction: api.Restriction{
					Scope: "a b c",
				},
				ExpiresAt: 100,
			},
		},
		okExp: true,
	}
	test.expected = test.wanted
	return test
}

func tightenCase2() tightenTestCase {
	test := tightenTestCase{
		name: "Empty Base, Multiple Wanted",
		base: Restrictions{},
		wanted: Restrictions{
			{
				Restriction: api.Restriction{
					Scope: "a b c",
				},
				ExpiresAt: 100,
			},
			{
				Restriction: api.Restriction{
					Scope: "d",
				},
				ExpiresAt: 100,
			},
		},
		okExp: true,
	}
	test.expected = test.wanted
	return test
}

func tightenCase3() tightenTestCase {
	test := tightenTestCase{
		name: "Wanted Empty",
		base: Restrictions{
			{
				Restriction: api.Restriction{
					Scope: "a b c",
				},
				ExpiresAt: 500,
			},
			{
				Restriction: api.Restriction{
					Scope: "a",
				},
				ExpiresAt: 1000,
			},
			{
				Restriction: api.Restriction{
					Scope: "d",
				},
				ExpiresAt: 50,
			},
		},
		wanted: Restrictions{},
		okExp:  true,
	}
	test.expected = test.base
	return test
}

func tightenCase4() tightenTestCase {
	test := tightenTestCase{
		name: "Restrict to One",
		base: Restrictions{
			{
				Restriction: api.Restriction{
					Scope: "a b c",
				},
				ExpiresAt: 500,
			},
			{
				Restriction: api.Restriction{
					Scope: "a",
				},
				ExpiresAt: 1000,
			},
			{
				Restriction: api.Restriction{
					Scope: "d",
				},
				ExpiresAt: 50,
			},
		},
		wanted: Restrictions{
			{
				Restriction: api.Restriction{
					Scope: "a b c",
				},
				ExpiresAt: 100,
			},
		},
		okExp: true,
	}
	test.expected = test.wanted
	return test
}

func tightenCase5() tightenTestCase {
	test := tightenTestCase{
		name: "Restrict to Two",
		base: Restrictions{
			{
				Restriction: api.Restriction{
					Scope: "a b c",
				},
				ExpiresAt: 500,
			},
			{
				Restriction: api.Restriction{
					Scope: "a",
				},
				ExpiresAt: 1000,
			},
			{
				Restriction: api.Restriction{
					Scope: "d",
				},
				ExpiresAt: 50,
			},
		},
		wanted: Restrictions{
			{
				Restriction: api.Restriction{
					Scope: "a b c",
				},
				ExpiresAt: 100,
			},
			{
				Restriction: api.Restriction{
					Scope: "d",
				},
				ExpiresAt: 50,
			},
		},
		okExp: true,
	}
	test.expected = test.wanted
	return test
}
func tightenCase6() tightenTestCase {
	test := tightenTestCase{
		name: "Restrict to Two 2",
		base: Restrictions{
			{
				Restriction: api.Restriction{
					Scope: "a b c",
				},
				ExpiresAt: 500,
			},
		},
		wanted: Restrictions{
			{
				Restriction: api.Restriction{
					Scope: "a c",
				},
				ExpiresAt: 100,
			},
			{
				Restriction: api.Restriction{
					Scope: "b",
				},
				ExpiresAt: 50,
			},
		},
		okExp: true,
	}
	test.expected = test.wanted
	return test
}
func tightenCase7() tightenTestCase {
	test := tightenTestCase{
		name: "Conflict",
		base: Restrictions{
			{
				Restriction: api.Restriction{
					Scope: "a b c",
				},
				ExpiresAt: 500,
			},
			{
				Restriction: api.Restriction{
					Scope: "a",
				},
				ExpiresAt: 1000,
			},
			{
				Restriction: api.Restriction{
					Scope: "d",
				},
				ExpiresAt: 50,
			},
		},
		wanted: Restrictions{
			{
				Restriction: api.Restriction{
					Scope: "a b c d",
				},
				ExpiresAt: 100,
			},
		},
		okExp: false,
	}
	test.expected = test.base
	return test
}

func tightenCase8() tightenTestCase {
	test := tightenTestCase{
		name: "Do not combine two clauses",
		base: Restrictions{
			{
				Restriction: api.Restriction{
					Scope: "a b c",
				},
				ExpiresAt: 500,
			},
			{
				Restriction: api.Restriction{
					Scope: "d",
				},
				ExpiresAt: 50,
			},
		},
		wanted: Restrictions{
			{
				Restriction: api.Restriction{
					Scope: "a b c d", // This is semantically different from base,
					// because it allows a token with all the scopes combined. One might want to not allow this.
				},
				ExpiresAt: 50,
			},
		},
		okExp: false,
	}
	test.expected = test.base
	return test
}

func tightenCase9() tightenTestCase {
	test := tightenTestCase{
		name: "Do not extend usages 1",
		base: Restrictions{
			{
				Restriction: api.Restriction{
					UsagesAT: utils.NewInt64(10),
				},
			},
		},
		wanted: Restrictions{
			{
				Restriction: api.Restriction{
					UsagesAT: utils.NewInt64(11),
				},
			},
		},
		okExp: false,
	}
	test.expected = test.base
	return test
}

func tightenCase10() tightenTestCase {
	test := tightenTestCase{
		name: "Do not extend usages 2",
		base: Restrictions{
			{
				Restriction: api.Restriction{
					UsagesAT: utils.NewInt64(10),
				},
			},
		},
		wanted: Restrictions{
			{
				Restriction: api.Restriction{
					UsagesAT: utils.NewInt64(4),
				},
			},
			{
				Restriction: api.Restriction{
					UsagesAT: utils.NewInt64(4),
				},
			},
			{
				Restriction: api.Restriction{
					UsagesAT: utils.NewInt64(4),
				},
			},
		},
		expected: Restrictions{
			{
				Restriction: api.Restriction{
					UsagesAT: utils.NewInt64(4),
				},
			},
			{
				Restriction: api.Restriction{
					UsagesAT: utils.NewInt64(4),
				},
			},
		},
		okExp: false,
	}
	return test
}

func tightenCase11() tightenTestCase {
	test := tightenTestCase{
		name: "Split Usages",
		base: Restrictions{
			{
				Restriction: api.Restriction{
					UsagesAT: utils.NewInt64(10),
				},
			},
		},
		wanted: Restrictions{
			{
				Restriction: api.Restriction{
					UsagesAT: utils.NewInt64(5),
				},
			},
			{
				Restriction: api.Restriction{
					UsagesAT: utils.NewInt64(3),
				},
			},
			{
				Restriction: api.Restriction{
					UsagesAT: utils.NewInt64(2),
				},
			},
		},
		okExp: true,
	}
	test.expected = test.wanted
	return test
}

func TestRestriction_isTighterThan(t *testing.T) {
	tests := []struct {
		name     string
		a        Restriction
		b        Restriction
		expected bool
	}{
		{
			name:     "Both Empty",
			a:        Restriction{},
			b:        Restriction{},
			expected: true,
		},
		{
			name:     "A Empty",
			a:        Restriction{},
			b:        Restriction{Restriction: api.Restriction{Scope: "some"}},
			expected: false,
		},
		{
			name:     "B Empty",
			a:        Restriction{Restriction: api.Restriction{Scope: "some"}},
			b:        Restriction{},
			expected: true,
		},
		{
			name:     "Not Before 1",
			a:        Restriction{NotBefore: 50},
			b:        Restriction{NotBefore: 100},
			expected: false,
		},
		{
			name:     "Not Before 2",
			a:        Restriction{NotBefore: 100},
			b:        Restriction{NotBefore: 50},
			expected: true,
		},
		{
			name:     "Not Before, A Empty",
			a:        Restriction{},
			b:        Restriction{NotBefore: 50},
			expected: false,
		},
		{
			name:     "Not Before, B Empty",
			a:        Restriction{NotBefore: 50},
			b:        Restriction{},
			expected: true,
		},
		{
			name:     "Expires At 1",
			a:        Restriction{ExpiresAt: 50},
			b:        Restriction{ExpiresAt: 100},
			expected: true,
		},
		{
			name:     "Expires At 2",
			a:        Restriction{ExpiresAt: 100},
			b:        Restriction{ExpiresAt: 50},
			expected: false,
		},
		{
			name:     "Expires At, A Empty",
			a:        Restriction{},
			b:        Restriction{ExpiresAt: 50},
			expected: false,
		},
		{
			name:     "Expires At, B Empty",
			a:        Restriction{ExpiresAt: 50},
			b:        Restriction{},
			expected: true,
		},
		{
			name:     "scope 1",
			a:        Restriction{Restriction: api.Restriction{Scope: "some scope"}},
			b:        Restriction{Restriction: api.Restriction{Scope: "some"}},
			expected: false,
		},
		{
			name:     "scope 2",
			a:        Restriction{Restriction: api.Restriction{Scope: "some"}},
			b:        Restriction{Restriction: api.Restriction{Scope: "some scope"}},
			expected: true,
		},
		{
			name:     "scope 3",
			a:        Restriction{Restriction: api.Restriction{Scope: "some scope"}},
			b:        Restriction{Restriction: api.Restriction{Scope: "some other"}},
			expected: false,
		},
		{
			name:     "scope 4",
			a:        Restriction{Restriction: api.Restriction{Scope: "some other"}},
			b:        Restriction{Restriction: api.Restriction{Scope: "some scope"}},
			expected: false,
		},
		{
			name:     "scope 5",
			a:        Restriction{Restriction: api.Restriction{Scope: "some other"}},
			b:        Restriction{Restriction: api.Restriction{Scope: "completely different"}},
			expected: false,
		},
		{
			name:     "scope A empty",
			a:        Restriction{},
			b:        Restriction{Restriction: api.Restriction{Scope: "some scopes"}},
			expected: false,
		},
		{
			name:     "scope B empty",
			a:        Restriction{Restriction: api.Restriction{Scope: "some scopes"}},
			b:        Restriction{},
			expected: true,
		},
		{
			name: "IP",
			a:    Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.12"}}},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"192.168.0.12",
						"192.168.0.14",
					},
				},
			},
			expected: true,
		},
		{
			name: "IP Reversed",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"192.168.0.12",
						"192.168.0.14",
					},
				},
			},
			b:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.12"}}},
			expected: false,
		},
		{
			name: "IP with explicit net",
			a:    Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.12/24"}}},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"192.168.0.12",
						"192.168.0.14",
					},
				},
			},
			expected: true,
		},
		{
			name: "IP with explicit net Reversed",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"192.168.0.12",
						"192.168.0.14",
					},
				},
			},
			b:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.12/24"}}},
			expected: false,
		},
		{
			name:     "IP Subnet",
			a:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.12"}}},
			b:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.0/24"}}},
			expected: true,
		},
		{
			name:     "IP Subnet Reversed",
			a:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.0/24"}}},
			b:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.12"}}},
			expected: false,
		},
		{
			name: "IP Subnet + IP",
			a:    Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.12"}}},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"192.168.0.0/24",
						"192.168.1.2",
					},
				},
			},
			expected: true,
		},
		{
			name: "IP Subnet + IP Reversed",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"192.168.0.0/24",
						"192.168.1.2",
					},
				},
			},
			b:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.12"}}},
			expected: false,
		},
		{
			name: "IP Subnets + IP",
			a:    Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.0/24"}}},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"192.168.0.0/24",
						"192.168.1.2",
					},
				},
			},
			expected: true,
		},
		{
			name: "IP Subnets + IP Reversed",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"192.168.0.0/24",
						"192.168.1.2",
					},
				},
			},
			b:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.0/24"}}},
			expected: false,
		},
		{
			name:     "IP Different sized Subnets",
			a:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.0/24"}}},
			b:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.0/16"}}},
			expected: true,
		},
		{
			name:     "IP Different sized Subnets Reversed",
			a:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.0/16"}}},
			b:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.0/24"}}},
			expected: false,
		},
		{
			name:     "IP Different sized Subnets 2",
			a:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.128.0/24"}}},
			b:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.0/16"}}},
			expected: true,
		},
		{
			name:     "IP Different sized Subnets 2 Reversed",
			a:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.0/16"}}},
			b:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.128.0/24"}}},
			expected: false,
		},
		{
			name:     "IP Different Subnets",
			a:        Restriction{Restriction: api.Restriction{Hosts: []string{"193.168.0.0/24"}}},
			b:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.0/16"}}},
			expected: false,
		},
		{
			name:     "IP Different Subnets Reversed",
			a:        Restriction{Restriction: api.Restriction{Hosts: []string{"192.168.0.0/16"}}},
			b:        Restriction{Restriction: api.Restriction{Hosts: []string{"193.168.0.0/24"}}},
			expected: false,
		},
		{
			name: "IP A empty",
			a:    Restriction{},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"192.168.0.12",
						"192.168.0.42",
					},
				},
			},
			expected: false,
		},
		{
			name: "IP B empty",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"192.168.0.12",
						"192.168.0.42",
					},
				},
			},
			b:        Restriction{},
			expected: true,
		},
		{
			name: "Hosts equal",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"test.example.com",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"test.example.com",
					},
				},
			},
			expected: true,
		},
		{
			name: "Hosts equal wildcard",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"*.example.com",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"*.example.com",
					},
				},
			},
			expected: true,
		},
		{
			name: "Hosts one wildcard",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"test.example.com",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"*.example.com",
					},
				},
			},
			expected: true,
		},
		{
			name: "Hosts one wildcard reversed",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"*.example.com",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"test.example.com",
					},
				},
			},
			expected: false,
		},
		{
			name: "Hosts two wildcards",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"*.test.example.com",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"*.example.com",
					},
				},
			},
			expected: true,
		},
		{
			name: "Hosts two wildcards reversed",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"*.example.com",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"*.test.example.com",
					},
				},
			},
			expected: false,
		},
		{
			name: "Hosts wildcard different",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"test.other.com",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"*.example.com",
					},
				},
			},
			expected: false,
		},
		{
			name: "Host with ip",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"stackoverflow.com",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"198.252.206.16",
					},
				},
			},
			expected: true,
		},
		{
			name: "Host with ip reversed",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"198.252.206.16",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"stackoverflow.com",
					},
				},
			},
			expected: true,
		},
		{
			name: "wildcard Host with ip",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"198.252.206.16",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"*stackoverflow.com",
					},
				},
			},
			expected: true,
		},
		{
			name: "wildcard Host with ip",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"198.252.206.16",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"*stackoverflow.com",
					},
				},
			},
			expected: true,
		},
		{
			name: "wildcard Host with ip reversed",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"*.stackoverflow.com",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"198.252.206.16",
					},
				},
			},
			expected: false,
		},
		{
			name: "wildcard Host 2 with ip",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"198.252.206.16",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"*.stackoverflow.com",
					},
				},
			},
			expected: false,
		},
		{
			name: "wildcard Host 2 with ip reversed",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"*.stackoverflow.com",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"198.252.206.16",
					},
				},
			},
			expected: false,
		},
		{
			name: "invalid ip for host",
			a: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"test.example.com",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Hosts: []string{
						"192.168.0.42",
					},
				},
			},
			expected: false,
		},
		{
			name: "GeoIP Allow Subset",
			a: Restriction{
				Restriction: api.Restriction{
					GeoIPAllow: []string{
						"de",
						"us",
					},
				},
			},
			b:        Restriction{Restriction: api.Restriction{GeoIPAllow: []string{"de"}}},
			expected: false,
		},
		{
			name: "GeoIP Allow Subset Reversed",
			a:    Restriction{Restriction: api.Restriction{GeoIPAllow: []string{"de"}}},
			b: Restriction{
				Restriction: api.Restriction{
					GeoIPAllow: []string{
						"de",
						"us",
					},
				},
			},
			expected: true,
		},
		{
			name: "GeoIP Allow Intersection",
			a: Restriction{
				Restriction: api.Restriction{
					GeoIPAllow: []string{
						"de",
						"fr",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					GeoIPAllow: []string{
						"de",
						"us",
					},
				},
			},
			expected: false,
		},
		{
			name: "GeoIP Allow Distinct",
			a: Restriction{
				Restriction: api.Restriction{
					GeoIPAllow: []string{
						"de",
						"fr",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					GeoIPAllow: []string{
						"jp",
						"us",
					},
				},
			},
			expected: false,
		},
		{
			name: "GeoIP Allow A Empty",
			a:    Restriction{Restriction: api.Restriction{}},
			b: Restriction{
				Restriction: api.Restriction{
					GeoIPAllow: []string{
						"jp",
						"us",
					},
				},
			},
			expected: false,
		},
		{
			name: "GeoIP Allow B Empty",
			a: Restriction{
				Restriction: api.Restriction{
					GeoIPAllow: []string{
						"jp",
						"us",
					},
				},
			},
			b:        Restriction{Restriction: api.Restriction{}},
			expected: true,
		},
		{
			name: "GeoIP Disallow Subset",
			a: Restriction{
				Restriction: api.Restriction{
					GeoIPDisallow: []string{
						"de",
						"us",
					},
				},
			},
			b:        Restriction{Restriction: api.Restriction{GeoIPDisallow: []string{"de"}}},
			expected: true,
		},
		{
			name: "GeoIP Disallow Subset Reversed",
			a:    Restriction{Restriction: api.Restriction{GeoIPDisallow: []string{"de"}}},
			b: Restriction{
				Restriction: api.Restriction{
					GeoIPDisallow: []string{
						"de",
						"us",
					},
				},
			},
			expected: false,
		},
		{
			name: "GeoIP Disallow Intersection",
			a: Restriction{
				Restriction: api.Restriction{
					GeoIPDisallow: []string{
						"de",
						"fr",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					GeoIPDisallow: []string{
						"de",
						"us",
					},
				},
			},
			expected: false,
		},
		{
			name: "GeoIP Disallow Distinct",
			a: Restriction{
				Restriction: api.Restriction{
					GeoIPDisallow: []string{
						"de",
						"fr",
					},
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					GeoIPDisallow: []string{
						"jp",
						"us",
					},
				},
			},
			expected: false,
		},
		{
			name: "GeoIP Disallow A Empty",
			a:    Restriction{Restriction: api.Restriction{}},
			b: Restriction{
				Restriction: api.Restriction{
					GeoIPDisallow: []string{
						"jp",
						"us",
					},
				},
			},
			expected: false,
		},
		{
			name: "GeoIP Disallow B Empty",
			a: Restriction{
				Restriction: api.Restriction{
					GeoIPDisallow: []string{
						"jp",
						"us",
					},
				},
			},
			b:        Restriction{Restriction: api.Restriction{}},
			expected: true,
		},
		{
			name:     "Usages AT",
			a:        Restriction{Restriction: api.Restriction{UsagesAT: utils.NewInt64(20)}},
			b:        Restriction{Restriction: api.Restriction{UsagesAT: utils.NewInt64(10)}},
			expected: false,
		},
		{
			name:     "Usages AT Reversed",
			a:        Restriction{Restriction: api.Restriction{UsagesAT: utils.NewInt64(10)}},
			b:        Restriction{Restriction: api.Restriction{UsagesAT: utils.NewInt64(20)}},
			expected: true,
		},
		{
			name:     "Usages AT A Empty",
			a:        Restriction{},
			b:        Restriction{Restriction: api.Restriction{UsagesAT: utils.NewInt64(20)}},
			expected: false,
		},
		{
			name:     "Usages AT B Empty",
			a:        Restriction{Restriction: api.Restriction{UsagesAT: utils.NewInt64(20)}},
			b:        Restriction{},
			expected: true,
		},
		{
			name:     "Usages Other",
			a:        Restriction{Restriction: api.Restriction{UsagesOther: utils.NewInt64(20)}},
			b:        Restriction{Restriction: api.Restriction{UsagesOther: utils.NewInt64(10)}},
			expected: false,
		},
		{
			name:     "Usages Other Reversed",
			a:        Restriction{Restriction: api.Restriction{UsagesOther: utils.NewInt64(10)}},
			b:        Restriction{Restriction: api.Restriction{UsagesOther: utils.NewInt64(20)}},
			expected: true,
		},
		{
			name:     "Usages Other A Empty",
			a:        Restriction{},
			b:        Restriction{Restriction: api.Restriction{UsagesOther: utils.NewInt64(20)}},
			expected: false,
		},
		{
			name:     "Usages Other B Empty",
			a:        Restriction{Restriction: api.Restriction{UsagesOther: utils.NewInt64(20)}},
			b:        Restriction{},
			expected: true,
		},
		{
			name: "Multiple A Empty",
			a:    Restriction{},
			b: Restriction{
				Restriction: api.Restriction{
					Scope:       "a",
					UsagesAT:    utils.NewInt64(50),
					UsagesOther: utils.NewInt64(100),
				},
			},
			expected: false,
		},
		{
			name: "Multiple B Empty",
			a: Restriction{
				Restriction: api.Restriction{
					Scope:       "a",
					UsagesAT:    utils.NewInt64(50),
					UsagesOther: utils.NewInt64(100),
				},
			},
			b:        Restriction{},
			expected: true,
		},
		{
			name: "Multiple 1",
			a: Restriction{
				Restriction: api.Restriction{
					Scope:       "a",
					UsagesAT:    utils.NewInt64(20),
					UsagesOther: utils.NewInt64(20),
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Scope:       "a b",
					UsagesAT:    utils.NewInt64(50),
					UsagesOther: utils.NewInt64(100),
				},
			},
			expected: true,
		},
		{
			name: "Multiple 1 Reversed",
			a: Restriction{
				Restriction: api.Restriction{
					Scope:       "a b",
					UsagesAT:    utils.NewInt64(50),
					UsagesOther: utils.NewInt64(100),
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Scope:       "a",
					UsagesAT:    utils.NewInt64(20),
					UsagesOther: utils.NewInt64(20),
				},
			},
			expected: false,
		},
		{
			name: "Multiple 2",
			a: Restriction{
				Restriction: api.Restriction{
					UsagesAT:    utils.NewInt64(100),
					UsagesOther: utils.NewInt64(50),
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					UsagesAT:    utils.NewInt64(50),
					UsagesOther: utils.NewInt64(100),
				},
			},
			expected: false,
		},
		{
			name: "Multiple 2 Reversed",
			a: Restriction{
				Restriction: api.Restriction{
					UsagesAT:    utils.NewInt64(50),
					UsagesOther: utils.NewInt64(100),
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					UsagesAT:    utils.NewInt64(100),
					UsagesOther: utils.NewInt64(50),
				},
			},
			expected: false,
		},
		{
			name: "Multiple Equal",
			a: Restriction{
				Restriction: api.Restriction{
					Scope:       "a b c",
					UsagesAT:    utils.NewInt64(20),
					UsagesOther: utils.NewInt64(20),
				},
			},
			b: Restriction{
				Restriction: api.Restriction{
					Scope:       "a b c",
					UsagesAT:    utils.NewInt64(20),
					UsagesOther: utils.NewInt64(20),
				},
			},
			expected: true,
		},
		{
			name: "All Equal",
			a: Restriction{
				NotBefore: 500,
				ExpiresAt: 1000,
				Restriction: api.Restriction{
					Scope: "a b c",
					Audiences: []string{
						"a",
						"b",
						"c",
					},
					Hosts: []string{
						"a",
						"b",
						"c",
					},
					GeoIPAllow: []string{
						"a",
						"b",
						"c",
					},
					GeoIPDisallow: []string{
						"a",
						"b",
						"c",
					},
					UsagesAT:    utils.NewInt64(20),
					UsagesOther: utils.NewInt64(20),
				},
			},
			b: Restriction{
				NotBefore: 500,
				ExpiresAt: 1000,
				Restriction: api.Restriction{
					Scope: "a b c",
					Audiences: []string{
						"a",
						"b",
						"c",
					},
					Hosts: []string{
						"a",
						"b",
						"c",
					},
					GeoIPAllow: []string{
						"a",
						"b",
						"c",
					},
					GeoIPDisallow: []string{
						"a",
						"b",
						"c",
					},
					UsagesAT:    utils.NewInt64(20),
					UsagesOther: utils.NewInt64(20),
				},
			},
			expected: true,
		},
		{
			name: "All",
			a: Restriction{
				NotBefore: 500,
				ExpiresAt: 1000,
				Restriction: api.Restriction{
					Scope: "a b c",
					Audiences: []string{
						"a",
						"b",
						"c",
					},
					Hosts: []string{
						"a",
						"b",
						"c",
					},
					GeoIPAllow: []string{
						"a",
						"b",
						"c",
					},
					GeoIPDisallow: []string{
						"a",
						"b",
						"c",
					},
					UsagesAT:    utils.NewInt64(20),
					UsagesOther: utils.NewInt64(20),
				},
			},
			b: Restriction{
				NotBefore: 700,
				ExpiresAt: 1000,
				Restriction: api.Restriction{
					Scope: "a b c",
					Audiences: []string{
						"a",
						"b",
						"c",
					},
					Hosts: []string{
						"a",
						"b",
						"c",
					},
					GeoIPAllow: []string{
						"a",
						"b",
						"c",
					},
					GeoIPDisallow: []string{
						"a",
						"b",
						"c",
					},
					UsagesAT:    utils.NewInt64(20),
					UsagesOther: utils.NewInt64(20),
				},
			},
			expected: false,
		},
		{
			name: "All Reversed",
			a: Restriction{
				NotBefore: 700,
				ExpiresAt: 1000,
				Restriction: api.Restriction{
					Scope: "a b c",
					Audiences: []string{
						"a",
						"b",
						"c",
					},
					Hosts: []string{
						"a",
						"b",
						"c",
					},
					GeoIPAllow: []string{
						"a",
						"b",
						"c",
					},
					GeoIPDisallow: []string{
						"a",
						"b",
						"c",
					},
					UsagesAT:    utils.NewInt64(20),
					UsagesOther: utils.NewInt64(20),
				},
			},
			b: Restriction{
				NotBefore: 500,
				ExpiresAt: 1000,
				Restriction: api.Restriction{
					Scope: "a b c",
					Audiences: []string{
						"a",
						"b",
						"c",
					},
					Hosts: []string{
						"a",
						"b",
						"c",
					},
					GeoIPAllow: []string{
						"a",
						"b",
						"c",
					},
					GeoIPDisallow: []string{
						"a",
						"b",
						"c",
					},
					UsagesAT:    utils.NewInt64(20),
					UsagesOther: utils.NewInt64(20),
				},
			},
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				tighter := test.a.isTighterThan(&test.b)
				if tighter != test.expected {
					if test.expected {
						t.Errorf("Actually '%+v' is tighter than '%+v'", test.a, test.b)
					} else {
						t.Errorf("Actually '%+v' is not tighter than '%+v'", test.a, test.b)
					}
				}
			},
		)
	}
}

func TestRestrictions_GetExpires(t *testing.T) {
	tests := []struct {
		name     string
		r        Restrictions
		expected unixtime.UnixTime
	}{
		{
			name:     "Empty",
			r:        Restrictions{},
			expected: 0,
		},
		{
			name:     "Infinite",
			r:        Restrictions{{ExpiresAt: 0}},
			expected: 0,
		},
		{
			name:     "One",
			r:        Restrictions{{ExpiresAt: 100}},
			expected: 100,
		},
		{
			name: "Multiple",
			r: Restrictions{
				{ExpiresAt: 100},
				{ExpiresAt: 300},
				{ExpiresAt: 500},
			},
			expected: 500,
		},
		{
			name: "Multiple and Infinite",
			r: Restrictions{
				{ExpiresAt: 100},
				{ExpiresAt: 300},
				{ExpiresAt: 0},
				{ExpiresAt: 500},
			},
			expected: 0,
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				expires := test.r.GetExpires()
				if test.expected != expires {
					t.Errorf("Expected %d, but got %d", test.expected, expires)
				}
			},
		)
	}
}

func TestRestrictions_GetNotBefore(t *testing.T) {
	tests := []struct {
		name     string
		r        Restrictions
		expected unixtime.UnixTime
	}{
		{
			name:     "Empty",
			r:        Restrictions{},
			expected: 0,
		},
		{
			name:     "Infinite",
			r:        Restrictions{{NotBefore: 0}},
			expected: 0,
		},
		{
			name:     "One",
			r:        Restrictions{{NotBefore: 100}},
			expected: 100,
		},
		{
			name: "Multiple",
			r: Restrictions{
				{NotBefore: 100},
				{NotBefore: 300},
				{NotBefore: 500},
			},
			expected: 100,
		},
		{
			name: "Multiple and Infinite",
			r: Restrictions{
				{NotBefore: 100},
				{NotBefore: 300},
				{NotBefore: 0},
				{NotBefore: 500},
			},
			expected: 0,
		},
	}
	for _, test := range tests {
		t.Run(
			test.name, func(t *testing.T) {
				nbf := test.r.GetNotBefore()
				if test.expected != nbf {
					t.Errorf("Expected %d, but got %d", test.expected, nbf)
				}
			},
		)
	}
}

func TestRestriction_hash(t *testing.T) {
	r := Restriction{
		NotBefore: 1599939600,
		ExpiresAt: 1599948600,
		Restriction: api.Restriction{
			Hosts:    []string{"192.168.0.31"},
			UsagesAT: utils.NewInt64(11),
		},
	}

	cases := []struct {
		name    string
		r       Restriction
		expHash string
	}{
		{
			name:    "hash",
			r:       r,
			expHash: "umVWTDmDgo2NLCgUWFtTn4G8PfjUZYTMxJ5QJGGo4ZNFOno76ggNdQcPhDsfWfYMy79n6oxnwiahbvHZUuI96w==",
		},
	}
	for _, c := range cases {
		t.Run(
			c.name, func(t *testing.T) {
				hash, err := c.r.hash()
				if err != nil {
					t.Error(err)
				}
				if string(hash) != c.expHash {
					t.Errorf("hash '%s' does not match expected hash '%s'", hash, c.expHash)
				}
			},
		)
	}
}

func TestRestriction_legacyHash(t *testing.T) {
	r := Restriction{
		NotBefore: 1599939600,
		ExpiresAt: 1599948600,
		Restriction: api.Restriction{
			Hosts:    []string{"192.168.0.31"},
			UsagesAT: utils.NewInt64(11),
		},
	}

	cases := []struct {
		name    string
		r       Restriction
		expHash string
	}{
		{
			name:    "legacy hash",
			r:       r,
			expHash: "BS3WfHbHNUiVU8sJ+F49H9+69HnFtfVDy2m22vBv588nZ0kGblVNxZEcrTN+5NUiRkM7W80N4VpPgwEZBZl+3g==",
		},
	}
	for _, c := range cases {
		t.Run(
			c.name, func(t *testing.T) {
				hash, err := c.r.legacyHash()
				if err != nil {
					t.Error(err)
				}
				if string(hash) != c.expHash {
					t.Errorf("hash '%s' does not match expected hash '%s'", hash, c.expHash)
				}
			},
		)
	}
}

func TestRestriction_VerifyTimeBased(t *testing.T) {
	now := unixtime.Now()
	cases := []struct {
		name string
		r    Restriction
		exp  bool
	}{
		{
			name: "Both 0",
			r: Restriction{
				NotBefore: 0,
				ExpiresAt: 0,
			},
			exp: true,
		},
		{
			name: "valid nbf",
			r: Restriction{
				NotBefore: now - 10,
				ExpiresAt: 0,
			},
			exp: true,
		},
		{
			name: "valid exp",
			r: Restriction{
				NotBefore: 0,
				ExpiresAt: now + 10,
			},
			exp: true,
		},
		{
			name: "invalid nbf",
			r: Restriction{
				NotBefore: now + 10,
				ExpiresAt: 0,
			},
			exp: false,
		},
		{
			name: "invalid exp",
			r: Restriction{
				NotBefore: 0,
				ExpiresAt: now - 10,
			},
			exp: false,
		},
		{
			name: "invalid nbf and exp",
			r: Restriction{
				NotBefore: now + 10,
				ExpiresAt: now - 10,
			},
			exp: false,
		},
	}
	for _, c := range cases {
		t.Run(
			c.name, func(t *testing.T) {
				valid := c.r.verifyTimeBased(log.StandardLogger())
				if valid != c.exp {
					t.Errorf(
						"For '%+v' expected time based attributes to verify as '%v' but got '%v'", c.r, c.exp, valid,
					)
				}
			},
		)
	}
}
