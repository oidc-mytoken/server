package pkg

import (
	"encoding/json"

	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
)

// WebRestrictions a type for representing restrictions.Restrictions in the consent screen
type WebRestrictions struct {
	restrictions.Restrictions
}

// Text returns a textual (json) representation of this WebRestrictions
func (r WebRestrictions) Text() string {
	data, _ := json.Marshal(r.Restrictions)
	return string(data)
}
