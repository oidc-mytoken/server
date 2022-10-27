package state

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/utils/hashutils"
)

const stateLen = 16

// State is a type for the oidc state
type State struct {
	state       string
	hash        string
	pollingCode string
}

// CreateState creates a new State and ConsentCode from the passed Info
func CreateState() (*State, *ConsentCode) {
	consentCode := NewConsentCode()
	s := consentCode.GetState()
	return NewState(s), consentCode
}

// NewState creates a new State from a state string
func NewState(state string) *State {
	return &State{
		state: state,
	}
}

// Hash returns the hash for this State
func (s *State) Hash() string {
	if s.hash == "" {
		s.hash = hashutils.SHA3_512Str([]byte(s.state))
	}
	return s.hash
}

// PollingCode returns the polling code for this State
func (s *State) PollingCode(rlog log.Ext1FieldLogger) string {
	if s.pollingCode == "" {
		s.pollingCode = hashutils.HMACBasedHash([]byte(s.state))[:config.Get().Features.Polling.Len]
		rlog.WithField("state", s.state).WithField(
			"polling_code", s.pollingCode,
		).Debug("Created polling_code for state")
	}
	return s.pollingCode
}

// State returns the state string for this State
func (s State) State() string {
	return s.state
}

// Value implements the driver.Valuer interface
func (s State) Value() (driver.Value, error) {
	return s.Hash(), nil
}

// Scan implements the sql.Scanner interface
func (s *State) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	ns := db.NewNullString("")
	if err := errors.WithStack(ns.Scan(src)); err != nil {
		return err
	}
	s.hash = ns.String
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (s *State) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &s.state)
	return errors.WithStack(err)
}

// MarshalJSON implements the json.Marshaler interface
func (s State) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(s.state)
	return data, errors.WithStack(err)
}
