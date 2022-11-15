package pkg

import (
	"encoding/json"

	"github.com/oidc-mytoken/utils/unixtime"
	"github.com/oidc-mytoken/utils/utils"

	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
)

// WebRestrictions a type for representing restrictions.Restrictions in the consent screen
type WebRestrictions struct {
	restrictions.Restrictions
	timeClass   *int
	ipClass     *bool
	scopeClass  *bool
	audClass    *bool
	usagesClass *bool
}

// Text returns a textual (json) representation of this WebRestrictions
func (r WebRestrictions) Text() string {
	data, _ := json.Marshal(r.Restrictions)
	return string(data)
}

func (r *WebRestrictions) getTimeClass() int {
	if r.timeClass != nil {
		return *r.timeClass
	}
	exp := r.GetExpires()
	if exp == 0 {
		r.timeClass = utils.NewInt(0)
	} else if exp-unixtime.Now() > 7*24*2600 {
		r.timeClass = utils.NewInt(1)
	} else {
		r.timeClass = utils.NewInt(2)
	}
	if r.timeClass != nil {
		return *r.timeClass
	}
	return -1
}

func (r *WebRestrictions) getScopeClass() bool {
	if r.scopeClass != nil {
		return *r.scopeClass
	}
	scopes := r.GetScopes()
	s := false
	if len(scopes) > 0 {
		s = true
	}
	r.scopeClass = &s
	return s
}

func (r *WebRestrictions) getIPClass() bool {
	if r.ipClass != nil {
		return *r.ipClass
	}
	ip := false
	for _, rr := range r.Restrictions {
		if len(rr.Hosts) > 0 {
			ip = true
			break
		}
		if len(rr.GeoIPAllow) > 0 {
			ip = true
			break
		}
		if len(rr.GeoIPDisallow) > 0 {
			ip = true
			break
		}
	}
	r.ipClass = &ip
	return ip
}

func (r *WebRestrictions) getAudClass() bool {
	if r.audClass != nil {
		return *r.audClass
	}
	auds := r.GetAudiences()
	a := false
	if len(auds) > 0 {
		a = true
	}
	r.audClass = &a
	return a
}

func (r *WebRestrictions) getUsageClass() bool {
	if r.usagesClass != nil {
		return *r.usagesClass
	}
	u := false
	for _, rr := range r.Restrictions {
		if rr.UsagesAT != nil {
			u = true
			break
		}
		if rr.UsagesOther != nil {
			u = true
			break
		}
	}
	r.usagesClass = &u
	return u
}

// TimeColorClass returns the html class for coloring the time dimension
func (r WebRestrictions) TimeColorClass() string {
	intClass := r.getTimeClass()
	switch intClass {
	case 0:
		return "text-danger"
	case 1:
		return "text-warning"
	case 2:
		return "text-success"
	default:
		return ""
	}
}

// TimeDescription returns a string describing the state of the time dimension
func (r WebRestrictions) TimeDescription() string {
	intClass := r.getTimeClass()
	switch intClass {
	case 0:
		return "This token has an infinite lifetime!"
	case 1:
		return "This token is long-lived."
	case 2:
		return "This token will expire within 7days."
	default:
		return ""
	}
}

// ScopeColorClass returns the html class for coloring the scope dimension
func (r WebRestrictions) ScopeColorClass() string {
	if r.getScopeClass() {
		return "text-success"
	}
	return "text-warning"
}

// ScopeDescription returns a string describing the state of the scope dimension
func (r WebRestrictions) ScopeDescription() string {
	if r.getScopeClass() {
		return "This token has restrictions for scopes."
	}
	return "This token can use all configured scopes."
}
