package restrictions

import (
	"database/sql/driver"
	"encoding/json"
	"math"
	"strings"

	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/utils/geoip"
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/utils"
	"github.com/oidc-mytoken/server/shared/utils/unixtime"
)

// Restrictions is a slice of Restriction
type Restrictions []Restriction

var eRKs model.RestrictionKeys

func disabledRestrictionKeys() model.RestrictionKeys {
	if eRKs == nil {
		eRKs = config.Get().Features.DisabledRestrictionKeys
	}
	return eRKs
}

// Restriction describes a token usage restriction
type Restriction struct {
	NotBefore       unixtime.UnixTime `json:"nbf,omitempty"`
	ExpiresAt       unixtime.UnixTime `json:"exp,omitempty"`
	api.Restriction `json:",inline"`
}

// ClearUnsupportedKeys sets default values for the keys that are not supported by this instance
func (r *Restrictions) ClearUnsupportedKeys() {
	for i, rr := range *r {
		if disabledRestrictionKeys().Has(model.RestrictionKeyNotBefore) {
			rr.NotBefore = 0
		}
		if disabledRestrictionKeys().Has(model.RestrictionKeyExpiresAt) {
			rr.ExpiresAt = 0
		}
		if disabledRestrictionKeys().Has(model.RestrictionKeyScope) {
			rr.Scope = ""
		}
		if disabledRestrictionKeys().Has(model.RestrictionKeyAudiences) {
			rr.Audiences = nil
		}
		if disabledRestrictionKeys().Has(model.RestrictionKeyIPs) {
			rr.IPs = nil
		}
		if disabledRestrictionKeys().Has(model.RestrictionKeyGeoIPAllow) {
			rr.GeoIPAllow = nil
		}
		if disabledRestrictionKeys().Has(model.RestrictionKeyGeoIPDisallow) {
			rr.GeoIPDisallow = nil
		}
		if disabledRestrictionKeys().Has(model.RestrictionKeyUsagesAT) {
			rr.UsagesAT = nil
		}
		if disabledRestrictionKeys().Has(model.RestrictionKeyUsagesOther) {
			rr.UsagesOther = nil
		}
		(*r)[i] = rr
	}
}

// hash returns the hash of this restriction
func (r *Restriction) hash() ([]byte, error) {
	j, err := json.Marshal(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return hashUtils.SHA512(j), nil
}

func (r *Restriction) verifyNbf(now unixtime.UnixTime) bool {
	if disabledRestrictionKeys().Has(model.RestrictionKeyNotBefore) {
		return true
	}
	return now >= r.NotBefore
}
func (r *Restriction) verifyExp(now unixtime.UnixTime) bool {
	if disabledRestrictionKeys().Has(model.RestrictionKeyExpiresAt) {
		return true
	}
	return r.ExpiresAt == 0 ||
		now <= r.ExpiresAt
}
func (r *Restriction) verifyTimeBased() bool {
	log.Trace("Verifying time based")
	now := unixtime.Now()
	return r.verifyNbf(now) && r.verifyExp(now)
}
func (r *Restriction) verifyIPBased(ip string) bool {
	return r.verifyIPs(ip) && r.verifyGeoIP(ip)
}
func (r *Restriction) verifyIPs(ip string) bool {
	if disabledRestrictionKeys().Has(model.RestrictionKeyIPs) {
		return true
	}
	log.Trace("Verifying ips")
	return len(r.IPs) == 0 ||
		utils.IPIsIn(ip, r.IPs)
}
func (r *Restriction) verifyGeoIP(ip string) bool {
	log.Trace("Verifying ip geo location")
	return r.verifyGeoIPDisallow(ip) && r.verifyGeoIPAllow(ip)
}
func (r *Restriction) verifyGeoIPAllow(ip string) bool {
	if disabledRestrictionKeys().Has(model.RestrictionKeyGeoIPAllow) {
		return true
	}
	log.Trace("Verifying ip geo location allow list")
	allow := r.GeoIPAllow
	if len(allow) == 0 {
		return true
	}
	return utils.StringInSlice(geoip.CountryCode(ip), allow)
}
func (r *Restriction) verifyGeoIPDisallow(ip string) bool {
	if disabledRestrictionKeys().Has(model.RestrictionKeyGeoIPDisallow) {
		return true
	}
	log.Trace("Verifying ip geo location disallow list")
	disallow := r.GeoIPDisallow
	if len(disallow) == 0 {
		return true
	}
	return !utils.StringInSlice(geoip.CountryCode(ip), disallow)
}
func (r *Restriction) getATUsageCounts(tx *sqlx.Tx, myID mtid.MTID) (*int64, error) {
	hash, err := r.hash()
	if err != nil {
		return nil, err
	}
	return mytokenrepohelper.GetTokenUsagesAT(tx, myID, string(hash))
}
func (r *Restriction) verifyATUsageCounts(tx *sqlx.Tx, myID mtid.MTID) bool {
	if disabledRestrictionKeys().Has(model.RestrictionKeyUsagesAT) {
		return true
	}
	log.Trace("Verifying AT usage count")
	if r.UsagesAT == nil {
		return true
	}
	usages, err := r.getATUsageCounts(tx, myID)
	if err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		return false
	}
	if usages == nil {
		//  was not used before
		log.WithFields(
			map[string]interface{}{
				"myID": myID.String(),
			},
		).Debug("Did not found restriction in database; it was not used before")
		return *r.UsagesAT > 0
	}
	log.WithFields(
		map[string]interface{}{
			"myID":       myID.String(),
			"used":       *usages,
			"usageLimit": *r.UsagesAT,
		},
	).Debug("Found restriction usage in db.")
	return *usages < *r.UsagesAT
}
func (r *Restriction) getOtherUsageCounts(tx *sqlx.Tx, myID mtid.MTID) (*int64, error) {
	hash, err := r.hash()
	if err != nil {
		return nil, err
	}
	return mytokenrepohelper.GetTokenUsagesOther(tx, myID, string(hash))
}
func (r *Restriction) verifyOtherUsageCounts(tx *sqlx.Tx, id mtid.MTID) bool {
	if disabledRestrictionKeys().Has(model.RestrictionKeyUsagesOther) {
		return true
	}
	log.Trace("Verifying other usage count")
	if r.UsagesOther == nil {
		return true
	}
	usages, err := r.getOtherUsageCounts(tx, id)
	if err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		return false
	}
	if usages == nil {
		// was not used before
		log.WithFields(
			map[string]interface{}{
				"id": id.String(),
			},
		).Debug("Did not found restriction in database; it was not used before")
		return *r.UsagesOther > 0
	}
	log.WithFields(
		map[string]interface{}{
			"id":         id.String(),
			"used":       *usages,
			"usageLimit": *r.UsagesAT,
		},
	).Debug("Found restriction usage in db.")
	return *usages < *r.UsagesOther
}
func (r *Restriction) verify(ip string) bool {
	return r.verifyTimeBased() &&
		r.verifyIPBased(ip)
}
func (r *Restriction) verifyAT(tx *sqlx.Tx, ip string, id mtid.MTID) bool {
	return r.verify(ip) && r.verifyATUsageCounts(tx, id)
}
func (r *Restriction) verifyOther(tx *sqlx.Tx, ip string, id mtid.MTID) bool {
	return r.verify(ip) &&
		r.verifyOtherUsageCounts(tx, id)
}

// UsedAT will update the usages_AT value for this restriction; it should be called after this restriction was used to
// obtain an access token;
func (r *Restriction) UsedAT(tx *sqlx.Tx, id mtid.MTID) error {
	js, err := json.Marshal(r)
	if err != nil {
		return errors.WithStack(err)
	}
	return mytokenrepohelper.IncreaseTokenUsageAT(tx, id, js)
}

// UsedOther will update the usages_other value for this restriction; it should be called after this restriction was
// used for other reasons than obtaining an access token;
func (r *Restriction) UsedOther(tx *sqlx.Tx, id mtid.MTID) error {
	js, err := json.Marshal(r)
	if err != nil {
		return errors.WithStack(err)
	}
	return mytokenrepohelper.IncreaseTokenUsageOther(tx, id, js)
}

// VerifyForAT verifies if this restrictions can be used to obtain an access token
func (r Restrictions) VerifyForAT(tx *sqlx.Tx, ip string, id mtid.MTID) bool {
	if len(r) == 0 {
		return true
	}
	return len(r.GetValidForAT(tx, ip, id)) > 0
}

// VerifyForOther verifies if this restrictions can be used for other actions than obtaining an access token
func (r Restrictions) VerifyForOther(tx *sqlx.Tx, ip string, id mtid.MTID) bool {
	if len(r) == 0 {
		return true
	}
	return len(r.GetValidForOther(tx, ip, id)) > 0
}

// GetValidForAT returns the subset of Restrictions that can be used to obtain an access token
func (r Restrictions) GetValidForAT(tx *sqlx.Tx, ip string, myID mtid.MTID) (ret Restrictions) {
	for _, rr := range r {
		if rr.verifyAT(tx, ip, myID) {
			log.Trace("Found a valid restriction")
			ret = append(ret, rr)
		}
	}
	return
}

// GetValidForOther returns the subset of Restrictions that can be used for other actions than obtaining an access token
func (r Restrictions) GetValidForOther(tx *sqlx.Tx, ip string, myID mtid.MTID) (ret Restrictions) {
	for _, rr := range r {
		if rr.verifyOther(tx, ip, myID) {
			ret = append(ret, rr)
		}
	}
	return
}

// WithScopes returns the subset of Restrictions that can be used with the specified scopes
func (r Restrictions) WithScopes(scopes []string) (ret Restrictions) {
	log.WithField("scopes", scopes).WithField("len", len(scopes)).Trace("Filter restrictions for scopes")
	if len(scopes) == 0 {
		log.Trace("scopes empty, returning all restrictions")
		return r
	}
	for _, rr := range r {
		if rr.Scope == "" || utils.IsSubSet(scopes, utils.SplitIgnoreEmpty(rr.Scope, " ")) {
			ret = append(ret, rr)
		}
	}
	return
}

// WithAudiences returns the subset of Restrictions that can be used with the specified audiences
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

// TokenUsages is a slice of TokenUsage
type TokenUsages []TokenUsage

// TokenUsage holds the information about the usages of an my token
type TokenUsage struct {
	MTID            string `db:"MT_id"`
	UsagesOtherUsed uint   `db:"usages_other"`
	UsagesATUsed    uint   `db:"usages_AT"`
}

// Scan implements the sql.Scanner interface.
func (r *Restrictions) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	val := src.([]uint8)
	return errors.WithStack(json.Unmarshal(val, &r))
}

// Value implements the driver.Valuer interface
func (r Restrictions) Value() (driver.Value, error) {
	if len(r) == 0 {
		return nil, nil
	}
	v, err := json.Marshal(r)
	return v, errors.WithStack(err)
}

// GetExpires gets the maximum (latest) expiration time of all restrictions
func (r *Restrictions) GetExpires() unixtime.UnixTime {
	if r == nil {
		return 0
	}
	exp := unixtime.UnixTime(0)
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

// GetNotBefore gets the minimal (earliest) nbf time of all restrictions
func (r *Restrictions) GetNotBefore() unixtime.UnixTime {
	if r == nil || len(*r) == 0 {
		return 0
	}
	nbf := unixtime.UnixTime(math.MaxInt64)
	for _, rr := range *r {
		if rr.NotBefore == 0 { // if one entry has no nbf the min nbf is 0
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

// SetMaxScopes sets the maximum scopes, i.e. all scopes are stripped from the restrictions if not included in the
// passed argument. This is used to eliminate requested scopes that are dropped by the provider. Don't use it to eliminate scopes that are not enabled for the oidc client, because it also could be a custom scope.
func (r *Restrictions) SetMaxScopes(mScopes []string) {
	for _, rr := range *r {
		rScopes := utils.SplitIgnoreEmpty(rr.Scope, " ")
		okScopes := utils.IntersectSlices(mScopes, rScopes)
		rr.Scope = strings.Join(okScopes, " ")
	}
}

// SetMaxAudiences sets the maximum audiences, i.e. all audiences are stripped from the restrictions if not included in
// the passed argument. This is used to eliminate requested audiences that are dropped by the provider.
func (r *Restrictions) SetMaxAudiences(mAud []string) {
	for _, rr := range *r {
		rr.Audiences = utils.IntersectSlices(mAud, rr.Audiences)
	}
}

// EnforceMaxLifetime enforces the maximum mytoken lifetime set by server admins. Returns true if the restrictions was
// changed.
func (r *Restrictions) EnforceMaxLifetime(issuer string) (changed bool) {
	maxLifetime := config.Get().ProviderByIssuer[issuer].MytokensMaxLifetime
	if maxLifetime == 0 {
		return
	}
	exp := unixtime.InSeconds(maxLifetime)
	if len(*r) == 0 {
		*r = append(*r, Restriction{ExpiresAt: exp})
		changed = true
		return
	}
	for i, rr := range *r {
		if rr.ExpiresAt == 0 || rr.ExpiresAt > exp {
			rr.ExpiresAt = exp
			(*r)[i] = rr
			changed = true
		}
	}
	return
}

// Tighten tightens/restricts a Restrictions with another set; if the wanted Restrictions are not tighter the original
// ones are returned
func Tighten(old, wanted Restrictions) (res Restrictions, ok bool) {
	if len(old) == 0 {
		ok = true
		res = wanted
		return
	}
	base := Restrictions{}
	if err := copier.CopyWithOption(&base, &old, copier.Option{DeepCopy: true}); err != nil {
		log.WithError(err).Error()
	}
	var droppedRestrictionsFromWanted bool
	for _, a := range wanted {
		thisOk := false
		for i, o := range base {
			if a.isTighterThan(o) {
				thisOk = true
				res = append(res, a)
				if o.UsagesOther != nil && a.UsagesOther != nil {
					*base[i].UsagesOther -= *a.UsagesOther
				}
				if o.UsagesAT != nil && a.UsagesAT != nil {
					*base[i].UsagesAT -= *a.UsagesAT
				}
				break
			}
		}
		if !thisOk {
			droppedRestrictionsFromWanted = true
		}
	}
	if len(res) == 0 { // if all from wanted are dropped, because they are not tighter than old, than use old
		res = old
		if len(wanted) == 0 { // the default for an empty restriction field is always using the parent's restrictions,
			// so this is fine
			ok = true
		}
		return
	}
	ok = !droppedRestrictionsFromWanted
	return
}

// ReplaceThisIp replaces the special value 'this' with the given ip.
func (r *Restrictions) ReplaceThisIp(ip string) {
	for _, rr := range *r {
		utils.ReplaceStringInSlice(&rr.IPs, "this", ip, false)
	}
}

func (r *Restrictions) removeIndex(i int) { // skipcq SCC-U1000
	copy((*r)[i:], (*r)[i+1:]) // Shift r[i+1:] left one index.
	*r = (*r)[:len(*r)-1]      // Truncate slice.
}

func (r *Restriction) isTighterThan(b Restriction) bool {
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
	if len(r.Audiences) == 0 && len(b.Audiences) > 0 || !utils.IsSubSet(
		r.Audiences, b.Audiences,
	) && len(b.Audiences) != 0 {
		return false
	}
	if len(r.IPs) == 0 && len(b.IPs) > 0 || !utils.IPsAreSubSet(r.IPs, b.IPs) && len(b.IPs) != 0 {
		return false
	}
	if len(r.GeoIPAllow) == 0 && len(b.GeoIPAllow) > 0 || !utils.IsSubSet(
		r.GeoIPAllow, b.GeoIPAllow,
	) && len(b.GeoIPAllow) != 0 {
		return false
	}
	if !utils.IsSubSet(
		b.GeoIPDisallow, r.GeoIPDisallow,
	) { // for Disallow-list r must have all the values from b to be tighter
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
