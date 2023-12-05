package profiled

import (
	"encoding/json"

	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/profilerepo"
)

// Rotation extends the api.Rotation with profile unmarshalling
type Rotation struct {
	api.Rotation
}

// MarshalJSON implements the json.Marshaler
func (p Rotation) MarshalJSON() ([]byte, error) {
	p.IncludedProfiles = nil
	return json.Marshal(p.Rotation)
}

// UnmarshalJSON implements the json.Marshaler interface
func (p *Rotation) UnmarshalJSON(bytes []byte) error {
	if bytes == nil {
		return nil
	}
	if string(bytes) == "null" {
		return nil
	}
	parser := profilerepo.NewDBProfileParser(log.StandardLogger())
	r, err := parser.ParseRotationTemplate(bytes)
	if r != nil {
		*p = Rotation{*r}
	}
	return err
}
