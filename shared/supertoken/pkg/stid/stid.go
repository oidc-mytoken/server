package stid

import (
	"database/sql/driver"
	"encoding/json"

	uuid "github.com/satori/go.uuid"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
)

// STID is a type for the super token id
type STID struct {
	uuid.UUID
	hash string
}

// New creates a new STID
func New() STID {
	return STID{
		UUID: uuid.NewV4(),
	}
}

func (i *STID) Valid() bool {
	return i.UUID.String() != "00000000-0000-0000-0000-000000000000"
}

func (i *STID) HashValid() bool {
	if len(i.hash) > 0 {
		return true
	}
	if i.UUID.String() == "00000000-0000-0000-0000-000000000000" {
		return false
	}
	return len(i.Hash()) > 0
}

func (i *STID) Hash() string {
	if len(i.hash) == 0 && i.Valid() {
		i.hash = hashUtils.SHA512Str(i.Bytes())
	}
	return i.hash
}

// Value implements the driver.Valuer interface
func (i STID) Value() (driver.Value, error) {
	ns := db.NewNullString(i.Hash())
	ns.Valid = i.HashValid()
	return ns.Value()
}

// Scan implements the sql.Scanner interface
func (i *STID) Scan(src interface{}) error {
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

func (i STID) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

func (i *STID) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &i.UUID)
}
