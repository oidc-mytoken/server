package restrictions

import (
	"encoding/json"
	"math"
	"strings"

	"github.com/zachmann/mytoken/internal/utils"
)

// Restrictions is a slice of Restriction
type Restrictions []Restriction

// Restriction describes a token usage restriction
type Restriction struct {
	NotBefore   int64    `json:"nbf,omitempty"`
	ExpiresAt   int64    `json:"exp,omitempty"`
	Scope       string   `json:"scope,omitempty"`
	Audiences   []string `json:"audience,omitempty"`
	IPs         []string `json:"ip,omitempty"`
	GeoIPWhite  []string `json:"geoip_white,omitempty"`
	GeoIPBlack  []string `json:"geoip_black,omitempty"`
	UsagesAT    uint     `json:"usages_AT,omitempty"`
	UsagesOther uint     `json:"usages_other,omitempty"`
}

type TokenUsages []TokenUsage

type TokenUsage struct {
	STID            string `db:"ST_id"`
	UsagesOtherUsed uint   `db:"usages_other"`
	UsagesATUsed    uint   `db:"usages_AT"`
}

// Scan unmarshals restrictions from a json db field
func (r *Restrictions) Scan(src interface{}) error {
	val := src.([]uint8)
	err := json.Unmarshal(val, &r)
	return err
}

// GetExpires gets the maximum (latest) expiration time of all restrictions
func (r *Restrictions) GetExpires() int64 {
	if r == nil {
		return 0
	}
	exp := int64(0)
	for _, rr := range *r {
		if rr.ExpiresAt == 0 { // if one entry has no expiry the max expiry is 0
			return 0
		}
		if rr.ExpiresAt > exp {
			exp = rr.ExpiresAt
		}
	}
	return exp
}

// GetNotBefore gets the minimal (earliest) notbefore time of all restrictions
func (r *Restrictions) GetNotBefore() int64 {
	if r == nil || len(*r) == 0 {
		return 0
	}
	nbf := int64(math.MaxInt64)
	for _, rr := range *r {
		if rr.NotBefore == 0 { // if one entry has no notbefore the min notbefore is 0
			return 0
		}
		if rr.NotBefore < nbf {
			nbf = rr.NotBefore
		}
	}
	return nbf
}

func Tighten(old, wanted Restrictions) (res Restrictions) {
	if len(old) == 0 {
		return wanted
	}
	for _, a := range wanted {
		for _, o := range old {
			if a.IsTighterThan(o) {
				res = append(res, a)
				break
			}
		}
	}
	if len(res) == 0 { // if all from wanted are dropped, because they are not tighter than old, than use old
		return old
	}
	return
}

func (r *Restrictions) RemoveIndex(i int) {
	copy((*r)[i:], (*r)[i+1:]) // Shift r[i+1:] left one index.
	// r[len(r)-1] = ""     // Erase last element (write zero value).
	*r = (*r)[:len(*r)-1] // Truncate slice.
}

func (r *Restriction) IsTighterThan(b Restriction) bool {
	if r.NotBefore < b.NotBefore {
		return false
	}
	if r.ExpiresAt == 0 && b.ExpiresAt != 0 || r.ExpiresAt > b.ExpiresAt && b.ExpiresAt != 0 {
		return false
	}
	rScopes := strings.Split(r.Scope, " ")
	if r.Scope == "" {
		rScopes = []string{}
	}
	bScopes := strings.Split(b.Scope, " ")
	if b.Scope == "" {
		bScopes = []string{}
	}
	if len(rScopes) == 0 && len(bScopes) > 0 || !utils.IsSubSet(rScopes, bScopes) && len(bScopes) != 0 {
		return false
	}
	if len(r.Audiences) == 0 && len(b.Audiences) > 0 || !utils.IsSubSet(r.Audiences, b.Audiences) && len(b.Audiences) != 0 {
		return false
	}
	if len(r.IPs) == 0 && len(b.IPs) > 0 || !utils.IPsAreSubSet(r.IPs, b.IPs) && len(b.IPs) != 0 {
		return false
	}
	if len(r.GeoIPWhite) == 0 && len(b.GeoIPWhite) > 0 || !utils.IsSubSet(r.GeoIPWhite, b.GeoIPWhite) && len(b.GeoIPWhite) != 0 {
		return false
	}
	if !utils.IsSubSet(b.GeoIPBlack, r.GeoIPBlack) { // for Blacklist r must have all the values from b to be tighter
		return false
	}
	if r.UsagesAT == 0 && b.UsagesAT != 0 || r.UsagesAT > b.UsagesAT && b.UsagesAT != 0 {
		return false
	}
	if r.UsagesOther == 0 && b.UsagesOther != 0 || r.UsagesOther > b.UsagesOther && b.UsagesOther != 0 {
		return false
	}
	return true
}
