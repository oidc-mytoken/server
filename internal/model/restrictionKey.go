package model

import (
	"encoding/json"

	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// RestrictionClaim is an enum like type for restriction keys
type RestrictionClaim int

// RestrictionClaims is a slice of RestrictionClaim
type RestrictionClaims []RestrictionClaim

// AllRestrictionClaimStrings holds all defined RestrictionClaim strings
var AllRestrictionClaimStrings = api.AllRestrictionClaims

// AllRestrictionClaims holds all defined RestrictionClaims
var AllRestrictionClaims RestrictionClaims

func init() {
	for i := 0; i < int(maxRestrictionClaim); i++ {
		AllRestrictionClaims = append(AllRestrictionClaims, RestrictionClaim(i))
	}
}

// RestrictionClaims
const ( // assert that these are in the same order as api.AllRestrictionKeys
	RestrictionClaimNotBefore RestrictionClaim = iota
	RestrictionClaimExpiresAt
	RestrictionClaimScope
	RestrictionClaimAudiences
	RestrictionClaimHosts
	RestrictionClaimGeoIPAllow
	RestrictionClaimGeoIPDisallow
	RestrictionClaimUsagesAT
	RestrictionClaimUsagesOther
	maxRestrictionClaim
)

// NewRestrictionClaim creates a new RestrictionClaim from the grant type string
func NewRestrictionClaim(s string) RestrictionClaim {
	for i, f := range AllRestrictionClaimStrings {
		if f == s {
			return RestrictionClaim(i)
		}
	}
	return -1
}

func (rc *RestrictionClaim) String() string {
	if *rc < 0 || int(*rc) >= len(AllRestrictionClaims) {
		return ""
	}
	return AllRestrictionClaimStrings[*rc]
}

// Valid checks that RestrictionClaim is a defined grant type
func (rc *RestrictionClaim) Valid() bool {
	return *rc < maxRestrictionClaim && *rc >= 0
}

const valueNotValidRestrFmt = "value '%s' not valid for RestrictionClaim"

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (rc *RestrictionClaim) UnmarshalYAML(value *yaml.Node) error {
	s := value.Value
	if s == "" {
		return errors.New("empty value in unmarshal grant type")
	}
	*rc = NewRestrictionClaim(s)
	if !rc.Valid() {
		return errors.Errorf(valueNotValidRestrFmt, s)
	}
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (rc *RestrictionClaim) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return errors.WithStack(err)
	}
	*rc = NewRestrictionClaim(s)
	if !rc.Valid() {
		return errors.Errorf(valueNotValidRestrFmt, s)
	}
	return nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (rc *RestrictionClaim) UnmarshalText(data []byte) error {
	s := string(data)
	*rc = NewRestrictionClaim(s)
	if !rc.Valid() {
		return errors.Errorf(valueNotValidRestrFmt, s)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (rc RestrictionClaim) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(rc.String())
	return data, errors.WithStack(err)
}

// Has checks if a RestrictionClaim is in a RestrictionClaims
func (rks RestrictionClaims) Has(rk RestrictionClaim) bool {
	for _, k := range rks {
		if k == rk {
			return true
		}
	}
	return false
}

// Disable subtracts the passed RestrictionClaims from this RestrictionClaims and returns the left RestrictionClaims
func (rks RestrictionClaims) Disable(disable RestrictionClaims) (left RestrictionClaims) {
	for _, r := range rks {
		if !disable.Has(r) {
			left = append(left, r)
		}
	}
	return
}
