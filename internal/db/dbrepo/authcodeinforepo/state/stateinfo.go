package state

import (
	"github.com/oidc-mytoken/server/internal/utils/singleasciiencode"
	pkgModel "github.com/oidc-mytoken/server/shared/model"
)

// Info is a type for holding the information encoded in a State
type Info struct {
	Native       bool
	ResponseType pkgModel.ResponseType
}

const infoAsciiLen = 2

// Encode encodes the Info into a string
func (i Info) Encode() string {
	fe := singleasciiencode.NewFlagEncoder()
	fe.Set("native", i.Native)
	flags := fe.Encode()
	responseType := singleasciiencode.EncodeNumber64(byte(i.ResponseType))
	return string([]byte{flags, responseType})
}

// Decode decodes the Info from a string
func (i *Info) Decode(s string) {
	length := len(s)
	if length < infoAsciiLen {
		return
	}
	responseType, _ := singleasciiencode.DecodeNumber64(s[length-1])
	flags := singleasciiencode.Decode(s[length-2], "native")
	i.ResponseType = pkgModel.ResponseType(responseType)
	i.Native, _ = flags.Get("native")
}

// CreateState creates a new State and ConsentCode from the passed Info
func CreateState(info Info) (*State, *ConsentCode) {
	consentCode := NewConsentCode(info)
	s := consentCode.GetState()
	return NewState(s), consentCode
}

// Parse parses a State and returns the encoded Info
func (s *State) Parse() (info Info) {
	info.Decode(s.State())
	return
}
