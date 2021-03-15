package transfercoderepo

import (
	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/shared/supertoken/pkg/stid"
)

// ShortToken holds database information of a short token
type ShortToken struct {
	proxyToken
}

// NewShortToken creates a new short token from the given jwt of a normal SuperToken
func NewShortToken(jwt string, mID stid.STID) (*ShortToken, error) {
	pt := newProxyToken(config.Get().Features.ShortTokens.Len)
	if err := pt.SetJWT(jwt, mID); err != nil {
		return nil, err
	}
	shortToken := &ShortToken{
		proxyToken: *pt,
	}
	return shortToken, nil
}

// ParseShortToken creates a new short token from a short token string
func ParseShortToken(token string) *ShortToken {
	return &ShortToken{proxyToken: *parseProxyToken(token)}
}
