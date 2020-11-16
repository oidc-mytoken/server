package capabilities

import (
	"database/sql/driver"
	"encoding/json"
)

// Constants for capabilities
const (
	CapabilityAT               Capability = "AT"
	CapabilityCreateST                    = "create_super_token"
	CapabilitySettings                    = "settings"
	CapabilityTokeninfoHistory            = "tokeninfo_history"
	CapabilityTokeninfoTree               = "tokeninfo_tree"
	CapabilityListST                      = "list_super_tokens"
)

// Capabilities is a slice of Capability
type Capabilities []Capability

// Capability is a capability string
type Capability string

// Scan implements the sql.Scanner interface.
func (c *Capabilities) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	val := src.([]uint8)
	err := json.Unmarshal(val, &c)
	return err
}

// Value implements the driver.Valuer interface
func (c Capabilities) Value() (driver.Value, error) {
	if len(c) == 0 {
		return nil, nil
	}
	return json.Marshal(c)
}

// Tighten tightens two set of Capabilities into one new
func Tighten(a, b Capabilities) (res Capabilities) {
	for _, bb := range b {
		for _, aa := range a {
			if bb == aa {
				res = append(res, bb)
			}
		}
	}
	return
}

func (c Capabilities) Has(a Capability) bool {
	for _, cc := range c {
		if cc == a {
			return true
		}
	}
	return false
}
