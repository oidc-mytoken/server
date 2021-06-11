package mtid

import (
	"database/sql/driver"
	"encoding/json"

	uuid "github.com/satori/go.uuid"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
)

// MTID is a type for the mytoken id
type MTID struct {
	uuid.UUID
	hash string
}

// New creates a new MTID
func New() MTID {
	return MTID{
		UUID: uuid.NewV4(),
	}
}

// Valid checks if the MTID is valid
func (i *MTID) Valid() bool {
	return i.UUID.String() != "00000000-0000-0000-0000-000000000000"
}

// HashValid checks if the MTID hash is valid
func (i *MTID) HashValid() bool {
	if i.hash != "" {
		return true
	}
	if i.UUID.String() == "00000000-0000-0000-0000-000000000000" {
		return false
	}
	return i.Hash() != ""
}

// Hash returns the MTID's hash
func (i *MTID) Hash() string {
	if i.hash == "" && i.Valid() {
		i.hash = hashUtils.SHA512Str(i.Bytes())
	}
	return i.hash
}

// Value implements the driver.Valuer interface
func (i MTID) Value() (driver.Value, error) {
	ns := db.NewNullString(i.Hash())
	ns.Valid = i.HashValid()
	return ns.Value()
}

// Scan implements the sql.Scanner interface
func (i *MTID) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	ns := db.NewNullString("")
	if err := ns.Scan(src); err != nil {
		return err
	}
	i.hash = ns.String
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (i MTID) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *MTID) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &i.UUID)
}
