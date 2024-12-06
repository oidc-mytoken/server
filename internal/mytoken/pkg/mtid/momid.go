package mtid

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// MOMID is a type for the mom-id (manage-other-mytokens)-id
type MOMID struct {
	MTID
}

// MarshalJSON implements the json.Marshaler interface
func (i MOMID) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(i.Hash())
	return data, errors.WithStack(err)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *MOMID) UnmarshalJSON(data []byte) error {
	return errors.WithStack(json.Unmarshal(data, &i.hash))
}
