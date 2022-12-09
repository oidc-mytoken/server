package profiled

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/profilerepo"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
)

// Restrictions extends the restrictions.Restrictions with profile unmarshalling
type Restrictions struct {
	restrictions.Restrictions
}

// MarshalJSON implements the json.Marshaler
func (p Restrictions) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Restrictions)
}

// UnmarshalJSON implements the json.Marshaler interface
func (p *Restrictions) UnmarshalJSON(bytes []byte) error {
	parser := profilerepo.NewDBProfileParser(log.StandardLogger())
	r, err := parser.ParseRestrictionsTemplate(bytes)
	if err != nil {
		return err
	}
	*p = Restrictions{restrictions.NewRestrictionsFromAPI(r)}
	return nil
}
