package mtid

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/utils/hashutils"
)

// MTID is a type for the mytoken id
type MTID struct {
	uuid.UUID
	hash string
}

// New creates a new MTID
func New() (MTID, error) {
	u, err := uuid.NewV4()
	return MTID{
		UUID: u,
	}, errors.WithStack(err)
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
		i.hash = hashutils.SHA512Str(i.Bytes())
	}
	return i.hash
}

// Value implements the driver.Valuer interface
func (i MTID) Value() (driver.Value, error) {
	ns := db.NewNullString(i.Hash())
	ns.Valid = i.HashValid()
	v, err := ns.Value()
	return v, errors.WithStack(err)
}

// Scan implements the sql.Scanner interface
func (i *MTID) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	ns := db.NewNullString("")
	if err := errors.WithStack(ns.Scan(src)); err != nil {
		return err
	}
	i.hash = ns.String
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (i MTID) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(i.String())
	return data, errors.WithStack(err)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *MTID) UnmarshalJSON(data []byte) error {
	return errors.WithStack(json.Unmarshal(data, &i.UUID))
}
