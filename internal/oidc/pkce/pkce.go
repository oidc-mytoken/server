package pkce

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/pkg/errors"
)

// PKCE is a type holding the information for a PKCE flow
type PKCE struct {
	verifier  string
	challenge string
	method    PKCEMethod
}

// PKCEMethod is a type for the code challenge methods
type PKCEMethod string

// Defines for the possible PKCEMethod
const (
	TransformationPlain = PKCEMethod("plain")
	TransformationS256  = PKCEMethod("S256")
)

func (m PKCEMethod) String() string {
	return string(m)
}

// NewPKCE creates a new PKCE for the passed verifier and PKCEMethod
func NewPKCE(verifier string, method PKCEMethod) *PKCE {
	return &PKCE{
		verifier: verifier,
		method:   method,
	}
}

// NewS256PKCE creates a new PKCE for the passed verifier and the PKCEMethod TransformationS256
func NewS256PKCE(verifier string) *PKCE {
	return NewPKCE(verifier, TransformationS256)
}

// Verifier returns the code_verifier
func (pkce PKCE) Verifier() string {
	return pkce.verifier
}

// Challenge returns the code_challenge according to the defined PKCEMethod
func (pkce *PKCE) Challenge() (string, error) {
	var err error
	if pkce.challenge == "" {
		pkce.challenge, err = pkce.transform()
	}
	return pkce.challenge, err
}

func (pkce PKCE) transform() (string, error) {
	switch pkce.method {
	case TransformationPlain:
		return pkce.plain(), nil
	case TransformationS256:
		return pkce.s256(), nil
	default:
		return "", errors.New("unknown code_challenge_method")
	}
}

func (pkce PKCE) plain() string {
	return pkce.verifier
}

func (pkce PKCE) s256() string {
	hash := sha256.Sum256([]byte(pkce.verifier))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
}
