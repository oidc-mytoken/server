package authcode

import "testing"

func TestParseState(t *testing.T) {
	stateInfos := []stateInfo{
		{Native: false},
		{Native: true},
	}
	for _, stateInfo := range stateInfos {
		state := createState(stateInfo)
		parsed := parseState(state)
		if stateInfo != parsed {
			t.Errorf("%+v was not correctly converted, instead got %+v from state '%s'", stateInfo, parsed, state)
		}
	}
}
