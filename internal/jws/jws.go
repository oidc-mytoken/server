package jws

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/pkg/errors"
	"github.com/zachmann/go-oidfed/pkg/jwk"

	"github.com/oidc-mytoken/server/internal/config"
)

// GenerateMytokenSigningKeyPair generates a cryptographic key pair for mytoken signing with the algorithm specified in
// the mytoken config.
func GenerateMytokenSigningKeyPair() (sk crypto.Signer, pk crypto.PublicKey, err error) {
	return generateKeyPair(config.Get().Signing.Mytoken.Alg, config.Get().Signing.Mytoken.RSAKeyLen)
}

// GenerateOIDCSigningKeyPair generates a cryptographic key pair for jwt signing within oidc communication with the
// algorithm specified in the mytoken config.
func GenerateOIDCSigningKeyPair() (sk crypto.Signer, pk crypto.PublicKey, err error) {
	return generateKeyPair(config.Get().Signing.OIDC.Alg, config.Get().Signing.OIDC.RSAKeyLen)
}

// GenerateFederationSigningKeyPair generates a cryptographic key pair for federation signing with the algorithm
// specified in the config.
func GenerateFederationSigningKeyPair() (sk crypto.Signer, pk crypto.PublicKey, err error) {
	return generateKeyPair(
		config.Get().Features.Federation.Signing.Alg, config.Get().Features.Federation.Signing.RSAKeyLen,
	)
}

// generateKeyPair generates a cryptographic key pair with the passed properties
func generateKeyPair(alg jwa.SignatureAlgorithm, rsaKeyLen int) (
	sk crypto.Signer, pk crypto.PublicKey,
	err error,
) {
	switch alg {
	case jwa.RS256, jwa.RS384, jwa.RS512, jwa.PS256, jwa.PS384, jwa.PS512:
		if rsaKeyLen <= 0 {
			return nil, nil, errors.Errorf("%s specified, but no valid RSA key len", alg)
		}
		sk, err = rsa.GenerateKey(rand.Reader, rsaKeyLen)
	case jwa.ES256:
		sk, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case jwa.ES384:
		sk, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case jwa.ES512:
		sk, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		err = errors.Errorf("unknown signing algorithm '%s'", alg)
		return
	}
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	pk = sk.Public()
	return
}

// ExportPrivateKeyAsPemStr exports the private key
func ExportPrivateKeyAsPemStr(sk crypto.Signer) string {
	switch sk := sk.(type) {
	case *rsa.PrivateKey:
		return exportRSAPrivateKeyAsPemStr(sk)
	case *ecdsa.PrivateKey:
		return exportECPrivateKeyAsPemStr(sk)
	default:
		return ""
	}
}

func exportECPrivateKeyAsPemStr(privkey *ecdsa.PrivateKey) string {
	privkeyBytes, _ := x509.MarshalECPrivateKey(privkey)
	privkeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: privkeyBytes,
		},
	)
	return string(privkeyPem)
}

func exportRSAPrivateKeyAsPemStr(privkey *rsa.PrivateKey) string {
	privkeyBytes := x509.MarshalPKCS1PrivateKey(privkey)
	privkeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkeyBytes,
		},
	)
	return string(privkeyPem)
}

// KeyUsage is a type indicating the usage purpose of a key
type KeyUsage string

// Predefined KeyUsage strings
const (
	KeyUsageMytokenSigning = KeyUsage("MT signing")
	KeyUsageFederation     = KeyUsage("oidcfed")
	KeyUsageOIDCSigning    = KeyUsage("oidc comm")
)

type signingKeys map[KeyUsage]signingKeyMaterial

type signingKeyMaterial struct {
	SK   crypto.Signer
	PK   crypto.PublicKey
	JWKS jwk.JWKS
}

var keys signingKeys

func init() {
	keys = make(signingKeys)
}

// GetSigningKey returns the private key
func GetSigningKey(usage KeyUsage) (sk crypto.Signer) {
	k, ok := keys[usage]
	if ok {
		sk = k.SK
	}
	return
}

// GetPublicKey returns the public key
func GetPublicKey(usage KeyUsage) (pk crypto.PublicKey) {
	k, ok := keys[usage]
	if ok {
		pk = k.PK
	}
	return
}

// GetJWKS returns the jwks
func GetJWKS(usage KeyUsage) (jwks jwk.JWKS) {
	k, ok := keys[usage]
	if ok {
		jwks = k.JWKS
	}
	return
}

// LoadMytokenSigningKey loads the private and public key for signing mytokens
func LoadMytokenSigningKey() {
	loadKey(config.Get().Signing.Mytoken.KeyFile, KeyUsageMytokenSigning, config.Get().Signing.Mytoken.Alg)
}

// LoadOIDCSigningKey loads the private and public key for signing operations within oidc communcation
func LoadOIDCSigningKey() {
	if config.Get().Signing.OIDC.KeyFile != "" {
		loadKey(config.Get().Signing.OIDC.KeyFile, KeyUsageOIDCSigning, config.Get().Signing.OIDC.Alg)
	}
}

// LoadFederationKey loads the private and public key for signing federation statements
func LoadFederationKey() {
	loadKey(
		config.Get().Features.Federation.Signing.KeyFile, KeyUsageFederation,
		config.Get().Features.Federation.Signing.Alg,
	)
}

// loadKey loads the private and public key from the passed keyfile
func loadKey(keyfile string, usage KeyUsage, alg jwa.SignatureAlgorithm) {
	keyFileContent, err := os.ReadFile(keyfile)
	if err != nil {
		panic(err)
	}
	var sk crypto.Signer
	switch alg {
	case jwa.RS256, jwa.RS384, jwa.RS512, jwa.PS256, jwa.PS384, jwa.PS512:
		sk, err = jwt.ParseRSAPrivateKeyFromPEM(keyFileContent)
		if err != nil {
			panic(err)
		}
	case jwa.ES256, jwa.ES384, jwa.ES512:
		sk, err = jwt.ParseECPrivateKeyFromPEM(keyFileContent)
		if err != nil {
			panic(err)
		}
	default:
		panic(fmt.Errorf("unknown signing alg"))
	}
	keyData, found := keys[usage]
	if !found {
		keyData = signingKeyMaterial{}
	}
	keyData.SK = sk
	keyData.PK = sk.Public()
	keyData.JWKS = jwk.KeyToJWKS(keyData.PK, alg)
	keys[usage] = keyData
}
