package state

import (
	"testing"

	"github.com/oidc-mytoken/server/pkg/model"
)

func TestParseState(t *testing.T) {
	stateInfos := []Info{
		{Native: false},
		{Native: true},
		{Native: false, ResponseType: model.ResponseTypeToken},
		{Native: true, ResponseType: model.ResponseTypeToken},
		{Native: false, ResponseType: model.ResponseTypeShortToken},
		{Native: true, ResponseType: model.ResponseTypeShortToken},
		{Native: false, ResponseType: model.ResponseTypeTransferCode},
		{Native: true, ResponseType: model.ResponseTypeTransferCode},
	}
	for _, stateInfo := range stateInfos {
		s, _ := CreateState(stateInfo)
		parsed := s.Parse()
		if stateInfo != parsed {
			t.Errorf("%+v was not correctly converted, instead got %+v from state '%+v'", stateInfo, parsed, s)
		}
	}
}
