package restrictions

import (
	"database/sql/driver"
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
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
	UsagesAT    *int64   `json:"usages_AT,omitempty"`
	UsagesOther *int64   `json:"usages_other,omitempty"`
	//Usages    *int64   `json:"usages,omitempty"`
}

func (r *Restriction) VerifyTimeBased() bool {
	now := time.Now().Unix()
	return (r.NotBefore == 0 || now >= r.NotBefore) &&
		now <= r.ExpiresAt
}
func (r *Restriction) VerifyIPBased(ip string) bool {
	return len(r.IPs) == 0 ||
		utils.IPIsIn(ip, r.IPs)
	//TODO check geoip
}
func (r *Restriction) VerifyATUsageCounts() bool {
	if r.UsagesAT == nil {
		return true
	}
	//TODO get usages from db and check
	return true
}
func (r *Restriction) VerifyOtherUsageCounts() bool {
	if r.UsagesOther == nil {
		return true
	}
	//TODO get usages from db and check
	return true
}
func (r *Restriction) verify(ip string) bool {
	return r.VerifyTimeBased() &&
		r.VerifyIPBased(ip)
}
func (r *Restriction) VerifyAT(ip string) bool {
	return r.verify(ip) &&
		r.VerifyATUsageCounts()
}
func (r *Restriction) VerifyOther(ip string) bool {
	return r.verify(ip) &&
		r.VerifyOtherUsageCounts()
}

func (r Restrictions) VerifyForAT(ip string) bool {
	if len(r) == 0 {
		return true
	}
	return len(r.GetValidForAT(ip)) > 0
}
func (r Restrictions) VerifyForOther(ip string) bool {
	if len(r) == 0 {
		return true
	}
	return len(r.GetValidForOther(ip)) > 0
}

func (r Restrictions) GetValidForAT(ip string) (ret Restrictions) {
	for _, rr := range r {
		if rr.VerifyAT(ip) {
			log.Trace("Found a valid restriction")
			ret = append(ret, rr)
		}
	}
	return
}
func (r Restrictions) GetValidForOther(ip string) (ret Restrictions) {
	for _, rr := range r {
		if rr.VerifyOther(ip) {
			ret = append(ret, rr)
		}
	}
	return
}

func (r Restrictions) WithScopes(scopes []string) (ret Restrictions) {
	log.WithField("scopes", scopes).WithField("len", len(scopes)).Trace("Filter restrictions for scopes")
	if len(scopes) == 0 {
		log.Trace("scopes empty, returning all restrictions")
		return r
	}
	for _, rr := range r {
		if len(rr.Scope) == 0 || utils.IsSubSet(scopes, utils.SplitIgnoreEmpty(rr.Scope, " ")) {
			ret = append(ret, rr)
		}
	}
	return
}
func (r Restrictions) WithAudiences(audiences []string) (ret Restrictions) {
	log.WithField("audiences", audiences).WithField("len", len(audiences)).Trace("Filter restrictions for audiences")
	if len(audiences) == 0 {
		log.Trace("audiences empty, returning all restrictions")
		return r
	}
	for _, rr := range r {
		if len(rr.Audiences) == 0 || utils.IsSubSet(audiences, rr.Audiences) {
			ret = append(ret, rr)
		}
	}
	return
}

type TokenUsages []TokenUsage

type TokenUsage struct {
	STID            string `db:"ST_id"`
	UsagesOtherUsed uint   `db:"usages_other"`
	UsagesATUsed    uint   `db:"usages_AT"`
}

// Scan implements the sql.Scanner interface.
func (r *Restrictions) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	val := src.([]uint8)
	err := json.Unmarshal(val, &r)
	return err
}

// Value implements the driver.Valuer interface
func (r Restrictions) Value() (driver.Value, error) {
	if len(r) == 0 {
		return nil, nil
	}
	return json.Marshal(r)
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

// GetScopes returns the union of all scopes, i.e. all scopes that must be requested at the issuer
func (r *Restrictions) GetScopes() (scopes []string) {
	for _, rr := range *r {
		scopes = append(scopes, utils.SplitIgnoreEmpty(rr.Scope, " ")...)
	}
	scopes = utils.UniqueSlice(scopes)
	return
}

// GetAudiences returns the union of all audiences, i.e. all audiences that must be requested at the issuer
func (r *Restrictions) GetAudiences() (auds []string) {
	for _, rr := range *r {
		auds = append(auds, rr.Audiences...)
	}
	auds = utils.UniqueSlice(auds)
	return
}

// SetMaxScopes sets the maximum scopes, i.e. all scopes are stripped from the restrictions if not included in the passed argument. This is used to eliminate requested scopes that are dropped by the provider. Don't use it to eliminate scopes that are not enabled for the oidc client, because it also could be a custom scope.
func (r *Restrictions) SetMaxScopes(mScopes []string) {
	for _, rr := range *r {
		rScopes := utils.SplitIgnoreEmpty(rr.Scope, " ")
		okScopes := utils.IntersectSlices(mScopes, rScopes)
		rr.Scope = strings.Join(okScopes, " ")
	}
}

// SetMaxAudiences sets the maximum audiences, i.e. all audiences are stripped from the restrictions if not included in the passed argument. This is used to eliminate requested audiences that are dropped by the provider.
func (r *Restrictions) SetMaxAudiences(mAud []string) {
	for _, rr := range *r {
		rr.Audiences = utils.IntersectSlices(mAud, rr.Audiences)
	}
}

func Tighten(old, wanted Restrictions) (res Restrictions) {
	if len(old) == 0 {
		return wanted
	}
	base := Restrictions{}
	copier.Copy(&base, &old)
	for i, a := range wanted {
		for _, o := range base {
			if a.IsTighterThan(o) {
				res = append(res, a)
				base = append(base[:i], base[i+1:]...)
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
	rScopes := utils.SplitIgnoreEmpty(r.Scope, " ")
	if r.Scope == "" {
		rScopes = []string{}
	}
	bScopes := utils.SplitIgnoreEmpty(b.Scope, " ")
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
	if utils.CompareNullableIntsWithNilAsInfinity(r.UsagesAT, b.UsagesAT) > 0 {
		return false
	}
	if utils.CompareNullableIntsWithNilAsInfinity(r.UsagesOther, b.UsagesOther) > 0 {
		return false
	}
	return true
}
