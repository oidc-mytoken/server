package restrictions

import (
	"database/sql/driver"
	"encoding/json"
	"math"
	"strings"

	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/unixtime"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	iutils "github.com/oidc-mytoken/server/internal/utils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/geoip"
	"github.com/oidc-mytoken/server/internal/utils/hashutils"
	"github.com/oidc-mytoken/server/internal/utils/iputils"
)

// Restrictions is a slice of Restriction
type Restrictions []*Restriction

var eRKs model.RestrictionClaims

func disabledRestrictionKeys() model.RestrictionClaims {
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

func NewRestrictionsFromAPI(apis api.Restrictions) (rs Restrictions) {
	for _, a := range apis {
		rs = append(
			rs, &Restriction{
				NotBefore:   unixtime.UnixTime(a.NotBefore),
				ExpiresAt:   unixtime.UnixTime(a.ExpiresAt),
				Restriction: *a,
			},
		)
	}
	return
}

// ClearUnsupportedKeys sets default values for the keys that are not supported by this instance
func (r *Restrictions) ClearUnsupportedKeys() {
	for _, rr := range *r {
		if disabledRestrictionKeys().Has(model.RestrictionClaimNotBefore) {
			rr.NotBefore = 0
		}
		if disabledRestrictionKeys().Has(model.RestrictionClaimExpiresAt) {
			rr.ExpiresAt = 0
		}
		if disabledRestrictionKeys().Has(model.RestrictionClaimScope) {
			rr.Scope = ""
		}
		if disabledRestrictionKeys().Has(model.RestrictionClaimAudiences) {
			rr.Audiences = nil
		}
		if disabledRestrictionKeys().Has(model.RestrictionClaimHosts) {
			rr.Hosts = nil
		}
		if disabledRestrictionKeys().Has(model.RestrictionClaimGeoIPAllow) {
			rr.GeoIPAllow = nil
		}
		if disabledRestrictionKeys().Has(model.RestrictionClaimGeoIPDisallow) {
			rr.GeoIPDisallow = nil
		}
		if disabledRestrictionKeys().Has(model.RestrictionClaimUsagesAT) {
			rr.UsagesAT = nil
		}
		if disabledRestrictionKeys().Has(model.RestrictionClaimUsagesOther) {
			rr.UsagesOther = nil
		}
		// (*r)[i] = rr
	}
}

// legacyHash returns the legacy hash of this restriction (using "ip" instead of "hosts" key
func (r *Restriction) legacyHash() ([]byte, error) {
	type legacy struct {
		NotBefore     unixtime.UnixTime `json:"nbf,omitempty"`
		ExpiresAt     unixtime.UnixTime `json:"exp,omitempty"`
		Scope         string            `json:"scope,omitempty"`
		Audiences     []string          `json:"audience,omitempty"`
		Hosts         []string          `json:"ip,omitempty"`
		GeoIPAllow    []string          `json:"geoip_allow,omitempty"`
		GeoIPDisallow []string          `json:"geoip_disallow,omitempty"`
		UsagesAT      *int64            `json:"usages_AT,omitempty"`
		UsagesOther   *int64            `json:"usages_other,omitempty"`
	}
	l := legacy{
		NotBefore:     r.NotBefore,
		ExpiresAt:     r.ExpiresAt,
		Scope:         r.Scope,
		Audiences:     r.Audiences,
		Hosts:         r.Hosts,
		GeoIPAllow:    r.GeoIPAllow,
		GeoIPDisallow: r.GeoIPDisallow,
		UsagesAT:      r.UsagesAT,
		UsagesOther:   r.UsagesOther,
	}
	j, err := json.Marshal(l)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return hashutils.SHA512(j), nil
}

// hash returns the hash of this restriction
func (r *Restriction) hash() ([]byte, error) {
	j, err := json.Marshal(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return hashutils.SHA512(j), nil
}

func (r *Restriction) verifyNbf(now unixtime.UnixTime) bool {
	if disabledRestrictionKeys().Has(model.RestrictionClaimNotBefore) {
		return true
	}
	return now >= r.NotBefore
}
func (r *Restriction) verifyExp(now unixtime.UnixTime) bool {
	if disabledRestrictionKeys().Has(model.RestrictionClaimExpiresAt) {
		return true
	}
	return r.ExpiresAt == 0 ||
		now <= r.ExpiresAt
}
func (r *Restriction) verifyTimeBased(rlog log.Ext1FieldLogger) bool {
	rlog.Trace("Verifying time based")
	now := unixtime.Now()
	return r.verifyNbf(now) && r.verifyExp(now)
}
func (r *Restriction) verifyLocationBased(rlog log.Ext1FieldLogger, ip string) bool {
	return r.verifyHosts(rlog, ip) && r.verifyGeoIP(rlog, ip)
}
func (r *Restriction) verifyHosts(rlog log.Ext1FieldLogger, ip string) bool {
	if disabledRestrictionKeys().Has(model.RestrictionClaimHosts) {
		return true
	}
	rlog.Trace("Verifying hosts")
	return len(r.Hosts) == 0 ||
		iputils.IPIsIn(ip, r.Hosts)
}
func (r *Restriction) verifyGeoIP(rlog log.Ext1FieldLogger, ip string) bool {
	rlog.Trace("Verifying ip geo location")
	return r.verifyGeoIPDisallow(rlog, ip) && r.verifyGeoIPAllow(rlog, ip)
}
func (r *Restriction) verifyGeoIPAllow(rlog log.Ext1FieldLogger, ip string) bool {
	if disabledRestrictionKeys().Has(model.RestrictionClaimGeoIPAllow) {
		return true
	}
	rlog.Trace("Verifying ip geo location allow list")
	allow := r.GeoIPAllow
	if len(allow) == 0 {
		return true
	}
	return utils.StringInSlice(geoip.CountryCode(ip), allow)
}
func (r *Restriction) verifyGeoIPDisallow(rlog log.Ext1FieldLogger, ip string) bool {
	if disabledRestrictionKeys().Has(model.RestrictionClaimGeoIPDisallow) {
		return true
	}
	rlog.Trace("Verifying ip geo location disallow list")
	disallow := r.GeoIPDisallow
	if len(disallow) == 0 {
		return true
	}
	return !utils.StringInSlice(geoip.CountryCode(ip), disallow)
}
func (r *Restriction) getATUsageCounts(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myID mtid.MTID) (*int64, error) {
	hash, err := r.hash()
	if err != nil {
		return nil, err
	}
	usages, err := mytokenrepohelper.GetTokenUsagesAT(rlog, tx, myID, string(hash))
	if err != nil || usages != nil {
		return usages, err
	}
	hash, err = r.legacyHash()
	if err != nil {
		return nil, err
	}
	return mytokenrepohelper.GetTokenUsagesAT(rlog, tx, myID, string(hash))
}
func (r *Restriction) verifyATUsageCounts(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myID mtid.MTID) bool {
	if disabledRestrictionKeys().Has(model.RestrictionClaimUsagesAT) {
		return true
	}
	rlog.Trace("Verifying AT usage count")
	if r.UsagesAT == nil {
		return true
	}
	usages, err := r.getATUsageCounts(rlog, tx, myID)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return false
	}
	if usages == nil {
		//  was not used before
		rlog.WithFields(
			map[string]interface{}{
				"myID": myID.String(),
			},
		).Debug("Did not found restriction in database; it was not used before")
		return *r.UsagesAT > 0
	}
	rlog.WithFields(
		map[string]interface{}{
			"myID":       myID.String(),
			"used":       *usages,
			"usageLimit": *r.UsagesAT,
		},
	).Debug("Found restriction usage in db.")
	return *usages < *r.UsagesAT
}
func (r *Restriction) getOtherUsageCounts(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myID mtid.MTID) (*int64, error) {
	hash, err := r.hash()
	if err != nil {
		return nil, err
	}
	return mytokenrepohelper.GetTokenUsagesOther(rlog, tx, myID, string(hash))
}
func (r *Restriction) verifyOtherUsageCounts(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID) bool {
	if disabledRestrictionKeys().Has(model.RestrictionClaimUsagesOther) {
		return true
	}
	rlog.Trace("Verifying other usage count")
	if r.UsagesOther == nil {
		return true
	}
	usages, err := r.getOtherUsageCounts(rlog, tx, id)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return false
	}
	if usages == nil {
		// was not used before
		rlog.WithFields(
			map[string]interface{}{
				"id": id.String(),
			},
		).Debug("Did not found restriction in database; it was not used before")
		return *r.UsagesOther > 0
	}
	rlog.WithFields(
		map[string]interface{}{
			"id":         id.String(),
			"used":       *usages,
			"usageLimit": *r.UsagesAT,
		},
	).Debug("Found restriction usage in db.")
	return *usages < *r.UsagesOther
}
func (r *Restriction) verify(rlog log.Ext1FieldLogger, ip string) bool {
	return r.verifyTimeBased(rlog) &&
		r.verifyLocationBased(rlog, ip)
}
func (r *Restriction) verifyAT(rlog log.Ext1FieldLogger, tx *sqlx.Tx, ip string, id mtid.MTID) bool {
	return r.verify(rlog, ip) && r.verifyATUsageCounts(rlog, tx, id)
}
func (r *Restriction) verifyOther(rlog log.Ext1FieldLogger, tx *sqlx.Tx, ip string, id mtid.MTID) bool {
	return r.verify(rlog, ip) &&
		r.verifyOtherUsageCounts(rlog, tx, id)
}

// UsedAT will update the usages_AT value for this restriction; it should be called after this restriction was used to
// obtain an access token;
func (r *Restriction) UsedAT(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID) error {
	js, err := json.Marshal(r)
	if err != nil {
		return errors.WithStack(err)
	}
	return mytokenrepohelper.IncreaseTokenUsageAT(rlog, tx, id, js)
}

// UsedOther will update the usages_other value for this restriction; it should be called after this restriction was
// used for other reasons than obtaining an access token;
func (r *Restriction) UsedOther(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID) error {
	js, err := json.Marshal(r)
	if err != nil {
		return errors.WithStack(err)
	}
	return mytokenrepohelper.IncreaseTokenUsageOther(rlog, tx, id, js)
}

// VerifyForAT verifies if this restrictions can be used to obtain an access token
func (r Restrictions) VerifyForAT(rlog log.Ext1FieldLogger, tx *sqlx.Tx, ip string, id mtid.MTID) bool {
	if len(r) == 0 {
		return true
	}
	return len(r.GetValidForAT(rlog, tx, ip, id)) > 0
}

// VerifyForOther verifies if this restrictions can be used for other actions than obtaining an access token
func (r Restrictions) VerifyForOther(rlog log.Ext1FieldLogger, tx *sqlx.Tx, ip string, id mtid.MTID) bool {
	if len(r) == 0 {
		return true
	}
	return len(r.GetValidForOther(rlog, tx, ip, id)) > 0
}

// GetValidForAT returns the subset of Restrictions that can be used to obtain an access token
func (r Restrictions) GetValidForAT(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, ip string, myID mtid.MTID,
) (ret Restrictions) {
	for _, rr := range r {
		if rr.verifyAT(rlog, tx, ip, myID) {
			rlog.Trace("Found a valid restriction")
			ret = append(ret, rr)
		}
	}
	return
}

// GetValidForOther returns the subset of Restrictions that can be used for other actions than obtaining an access token
func (r Restrictions) GetValidForOther(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, ip string, myID mtid.MTID,
) (ret Restrictions) {
	for _, rr := range r {
		if rr.verifyOther(rlog, tx, ip, myID) {
			ret = append(ret, rr)
		}
	}
	return
}

// WithScopes returns the subset of Restrictions that can be used with the specified scopes
func (r Restrictions) WithScopes(rlog log.Ext1FieldLogger, scopes []string) (ret Restrictions) {
	rlog.WithField("scopes", scopes).WithField("len", len(scopes)).Trace("Filter restrictions for scopes")
	if len(scopes) == 0 {
		rlog.Trace("scopes empty, returning all restrictions")
		return r
	}
	for _, rr := range r {
		if rr.Scope == "" || utils.IsSubSet(scopes, iutils.SplitIgnoreEmpty(rr.Scope, " ")) {
			ret = append(ret, rr)
		}
	}
	return
}

// WithAudiences returns the subset of Restrictions that can be used with the specified audiences
func (r Restrictions) WithAudiences(rlog log.Ext1FieldLogger, audiences []string) (ret Restrictions) {
	rlog.WithField("audiences", audiences).WithField("len", len(audiences)).Trace("Filter restrictions for audiences")
	if len(audiences) == 0 {
		rlog.Trace("audiences empty, returning all restrictions")
		return r
	}
	for _, rr := range r {
		if len(rr.Audiences) == 0 || utils.IsSubSet(audiences, rr.Audiences) {
			ret = append(ret, rr)
		}
	}
	return
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
		scopes = append(scopes, iutils.SplitIgnoreEmpty(rr.Scope, " ")...)
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
// passed argument. This is used to eliminate requested scopes that are dropped by the provider. Don't use it to
// eliminate scopes that are not enabled for the oidc client, because it also could be a custom scope.
func (r *Restrictions) SetMaxScopes(mScopes []string) {
	for _, rr := range *r {
		rScopes := iutils.SplitIgnoreEmpty(rr.Scope, " ")
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
		*r = append(*r, &Restriction{ExpiresAt: exp})
		changed = true
		return
	}
	for _, rr := range *r {
		if rr.ExpiresAt == 0 || rr.ExpiresAt > exp {
			rr.ExpiresAt = exp
			// (*r)[i] = rr
			changed = true
		}
	}
	return
}

// Tighten tightens/restricts a Restrictions with another set; if the wanted Restrictions are not tighter the original
// ones are returned
func Tighten(rlog log.Ext1FieldLogger, old, wanted Restrictions) (res Restrictions, ok bool) {
	if len(old) == 0 {
		ok = true
		res = wanted
		return
	}
	base := Restrictions{}
	if err := copier.CopyWithOption(&base, &old, copier.Option{DeepCopy: true}); err != nil {
		rlog.WithError(err).Error()
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

// ReplaceThisIP replaces the special value 'this' with the given ip.
func (r *Restrictions) ReplaceThisIP(ip string) {
	for _, rr := range *r {
		utils.ReplaceStringInSlice(&rr.Hosts, "this", ip, false)
	}
}

func (r *Restriction) isTighterThan(b *Restriction) bool {
	if r.NotBefore < b.NotBefore {
		return false
	}
	if r.ExpiresAt == 0 && b.ExpiresAt != 0 || r.ExpiresAt > b.ExpiresAt && b.ExpiresAt != 0 {
		return false
	}
	rScopes := iutils.SplitIgnoreEmpty(r.Scope, " ")
	if r.Scope == "" {
		rScopes = []string{}
	}
	bScopes := iutils.SplitIgnoreEmpty(b.Scope, " ")
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
	if len(r.Hosts) == 0 && len(b.Hosts) > 0 || !iputils.IPsAreSubSet(r.Hosts, b.Hosts) && len(b.Hosts) != 0 {
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
	if iutils.CompareNullableIntsWithNilAsInfinity(r.UsagesAT, b.UsagesAT) > 0 {
		return false
	}
	if iutils.CompareNullableIntsWithNilAsInfinity(r.UsagesOther, b.UsagesOther) > 0 {
		return false
	}
	return true
}
