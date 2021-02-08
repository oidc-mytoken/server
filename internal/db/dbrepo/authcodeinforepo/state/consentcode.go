package state

import (
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
	"github.com/oidc-mytoken/server/shared/utils"
)

const stateLen = 16
const consentCodeLen = 8

func NewConsentCode(info Info) *ConsentCode {
	return &ConsentCode{
		r:           utils.RandASCIIString(consentCodeLen),
		encodedInfo: info.Encode(),
	}
}

func ParseConsentCode(cc string) *ConsentCode {
	return &ConsentCode{
		r:           cc[:len(cc)-infoAsciiLen],
		encodedInfo: cc[len(cc)-infoAsciiLen:],
		public:      cc,
	}
}

type ConsentCode struct {
	r           string
	encodedInfo string
	public      string
	state       string
}

func (c *ConsentCode) String() string {
	if len(c.public) > 0 {
		return c.public
	}
	c.public = c.r + c.encodedInfo
	return c.public
}

func (c *ConsentCode) GetState() string {
	if len(c.state) > 0 {
		return c.state
	}
	c.state = hashUtils.HMACSHA512Str([]byte(c.r), []byte("state"))[:stateLen] + c.encodedInfo
	return c.state
}
