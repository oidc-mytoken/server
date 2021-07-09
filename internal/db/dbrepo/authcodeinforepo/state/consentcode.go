package state

import (
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
	"github.com/oidc-mytoken/server/shared/utils"
)

const stateLen = 16
const consentCodeLen = 8

// NewConsentCode creates a new ConsentCode
func NewConsentCode(info Info) *ConsentCode {
	return &ConsentCode{
		r:           utils.RandASCIIString(consentCodeLen),
		encodedInfo: info.Encode(),
	}
}

// ParseConsentCode parses a string into a ConsentCode
func ParseConsentCode(cc string) *ConsentCode {
	return &ConsentCode{
		r:           cc[:len(cc)-infoAsciiLen],
		encodedInfo: cc[len(cc)-infoAsciiLen:],
		public:      cc,
	}
}

// ConsentCode is type for the code used for giving consent to mytoken
type ConsentCode struct {
	r           string
	encodedInfo string
	public      string
	state       string
}

func (c *ConsentCode) String() string {
	if c.public == "" {
		c.public = c.r + c.encodedInfo
	}
	return c.public
}

// GetState returns the state linked to a ConsentCode
func (c *ConsentCode) GetState() string {
	if c.state == "" {
		c.state = hashUtils.HMACSHA3Str([]byte("state"), []byte(c.r))[:stateLen] + c.encodedInfo
	}
	return c.state
}
