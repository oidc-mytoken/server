package profiled

import (
	"encoding/json"

	"github.com/oidc-mytoken/api/v0"

	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/profilerepo"
)

// Capabilities extends the api.Capabilities with profile unmarshalling
type Capabilities struct {
	api.Capabilities
}

// MarshalJSON implements the json.Marshaler
func (p Capabilities) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Capabilities)
}

// UnmarshalJSON implements the json.Marshaler interface
func (p *Capabilities) UnmarshalJSON(bytes []byte) error {
	if bytes == nil || string(bytes) == "null" {
		return nil
	}
	parser := profilerepo.NewDBProfileParser(log.StandardLogger())
	c, err := parser.ParseCapabilityTemplate(bytes)
	if err != nil {
		return err
	}
	*p = Capabilities{c}
	return nil
}
