package model

import (
	"encoding/json"
	"fmt"

	"github.com/oidc-mytoken/api/v0"
	yaml "gopkg.in/yaml.v3"
)

// TokeninfoAction is an enum like type for tokeninfo actions
type TokeninfoAction int

// AllTokeninfoActions holds all defined TokenInfo strings
var AllTokeninfoActions = api.AllTokeninfoActions

// TokeninfoActions
const ( // assert that these are in the same order as api.AllTokeninfoActions
	TokeninfoActionIntrospect TokeninfoAction = iota
	TokeninfoActionEventHistory
	TokeninfoActionSubtokenTree
	TokeninfoActionListMytokens
	maxTokeninfoAction
)

// NewTokeninfoAction creates a new TokeninfoAction from the tokeninfo action string
func NewTokeninfoAction(s string) TokeninfoAction {
	for i, f := range AllTokeninfoActions {
		if f == s {
			return TokeninfoAction(i)
		}
	}
	return -1
}

func (a *TokeninfoAction) String() string {
	if *a < 0 || int(*a) >= len(AllTokeninfoActions) {
		return ""
	}
	return AllTokeninfoActions[*a]
}

// Valid checks that TokeninfoAction is a defined tokeninfo action
func (a *TokeninfoAction) Valid() bool {
	return *a < maxTokeninfoAction && *a >= 0
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (a *TokeninfoAction) UnmarshalYAML(value *yaml.Node) error {
	s := value.Value
	if s == "" {
		return fmt.Errorf("empty value in unmarshal TokeninfoAction")
	}
	*a = NewTokeninfoAction(s)
	if !a.Valid() {
		return fmt.Errorf("value '%s' not valid for TokeninfoAction", s)
	}
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (a *TokeninfoAction) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*a = NewTokeninfoAction(s)
	if !a.Valid() {
		return fmt.Errorf("value '%s' not valid for TokeninfoAction", s)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (a TokeninfoAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

// AddToSliceIfNotFound adds the TokeninfoAction to a slice s if it is not already there
func (a TokeninfoAction) AddToSliceIfNotFound(s *[]TokeninfoAction) {
	for _, ss := range *s {
		if ss == a {
			return
		}
	}
	*s = append(*s, a)
}
