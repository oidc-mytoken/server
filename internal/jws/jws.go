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
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/config"
)

// GenerateMytokenSigningKeyPair generates a cryptographic key pair for mytoken signing with the algorithm specified in
// the mytoken config.
func GenerateMytokenSigningKeyPair() (sk crypto.Signer, pk crypto.PublicKey, err error) {
	return generateKeyPair(config.Get().Signing.Alg, config.Get().Signing.RSAKeyLen)
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
)

type signingKeys map[KeyUsage]signingKeyMaterial

type signingKeyMaterial struct {
	SK   crypto.Signer
	PK   crypto.PublicKey
	JWKS jwk.Set
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
func GetJWKS(usage KeyUsage) (jwks jwk.Set) {
	k, ok := keys[usage]
	if ok {
		jwks = k.JWKS
	}
	return
}

// LoadMytokenSigningKey loads the private and public key for signing mytokens
func LoadMytokenSigningKey() {
	loadKey(config.Get().Signing.KeyFile, KeyUsageMytokenSigning, config.Get().Signing.Alg)
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
		keyData = signingKeyMaterial{
			JWKS: jwk.NewSet(),
		}
	}
	keyData.SK = sk
	keyData.PK = sk.Public()
	key, err := jwk.New(keyData.PK)
	if err != nil {
		panic(err)
	}
	if err = jwk.AssignKeyID(key); err != nil {
		panic(err)
	}
	if err = key.Set(jwk.KeyUsageKey, jwk.ForSignature); err != nil {
		panic(err)
	}
	if err = key.Set(jwk.AlgorithmKey, alg); err != nil {
		panic(err)
	}
	keyData.JWKS.Add(key)
	keys[usage] = keyData
}
