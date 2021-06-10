package model

import (
	"encoding/json"
	"fmt"

	"github.com/oidc-mytoken/api/v0"
	yaml "gopkg.in/yaml.v3"
)

// RestrictionKey is an enum like type for restriction keys
type RestrictionKey int

// RestrictionKeys is a slice of RestrictionKey
type RestrictionKeys []RestrictionKey

// AllRestrictionKeyStrings holds all defined RestrictionKey strings
var AllRestrictionKeyStrings = api.AllRestrictionKeys

var AllRestrictionKeys RestrictionKeys

func init() {
	for i := 0; i < int(maxRestrictionKey); i++ {
		AllRestrictionKeys = append(AllRestrictionKeys, RestrictionKey(i))
	}
}

// RestrictionKeys
const ( // assert that these are in the same order as api.AllRestrictionKeys
	RestrictionKeyNotBefore RestrictionKey = iota
	RestrictionKeyExpiresAt
	RestrictionKeyScope
	RestrictionKeyAudiences
	RestrictionKeyIPs
	RestrictionKeyGeoIPAllow
	RestrictionKeyGeoIPDisallow
	RestrictionKeyUsagesAT
	RestrictionKeyUsagesOther
	maxRestrictionKey
)

// NewRestrictionKey creates a new RestrictionKey from the grant type string
func NewRestrictionKey(s string) RestrictionKey {
	for i, f := range AllRestrictionKeyStrings {
		if f == s {
			return RestrictionKey(i)
		}
	}
	return -1
}

func (rk *RestrictionKey) String() string {
	if *rk < 0 || int(*rk) >= len(AllRestrictionKeys) {
		return ""
	}
	return AllRestrictionKeyStrings[*rk]
}

// Valid checks that RestrictionKey is a defined grant type
func (rk *RestrictionKey) Valid() bool {
	return *rk < maxRestrictionKey && *rk >= 0
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (rk *RestrictionKey) UnmarshalYAML(value *yaml.Node) error {
	s := value.Value
	if s == "" {
		return fmt.Errorf("empty value in unmarshal grant type")
	}
	*rk = NewRestrictionKey(s)
	if !rk.Valid() {
		return fmt.Errorf("value '%s' not valid for RestrictionKey", s)
	}
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (rk *RestrictionKey) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*rk = NewRestrictionKey(s)
	if !rk.Valid() {
		return fmt.Errorf("value '%s' not valid for RestrictionKey", s)
	}
	return nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (rk *RestrictionKey) UnmarshalText(data []byte) error {
	s := string(data)
	*rk = NewRestrictionKey(s)
	if !rk.Valid() {
		return fmt.Errorf("value '%s' not valid for RestrictionKey", s)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (rk RestrictionKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(rk.String())
}

// Has checks if a a RestrictionKey is in a RestrictionKeys
func (rks RestrictionKeys) Has(rk RestrictionKey) bool {
	for _, k := range rks {
		if k == rk {
			return true
		}
	}
	return false
}

func (rks RestrictionKeys) Disable(disable RestrictionKeys) (left RestrictionKeys) {
	for _, r := range rks {
		if !disable.Has(r) {
			left = append(left, r)
		}
	}
	return
}
