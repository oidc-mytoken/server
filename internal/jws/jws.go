package jws

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/coreos/go-oidc/v3/oidc"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/config"
)

// GenerateKeyPair generates a cryptographic key pair for the algorithm specified in the mytoken config.
func GenerateKeyPair() (sk, pk interface{}, err error) {
	alg := config.Get().Signing.Alg
	switch alg {
	case oidc.RS256, oidc.RS384, oidc.RS512, oidc.PS256, oidc.PS384, oidc.PS512:
		keyLen := config.Get().Signing.RSAKeyLen
		if keyLen <= 0 {
			return nil, nil, errors.Errorf("%s specified, but no valid RSA key len", alg)
		}
		sk, err = rsa.GenerateKey(rand.Reader, keyLen)
	case oidc.ES256:
		sk, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case oidc.ES384:
		sk, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case oidc.ES512:
		sk, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return nil, nil, errors.Errorf("unknown signing algorithm '%s'", alg)
	}
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	switch sk := sk.(type) {
	case *rsa.PrivateKey:
		pk = &sk.PublicKey
	case *ecdsa.PrivateKey:
		pk = &sk.PublicKey
	default:
		err = errors.Errorf("something went wrong, we just created an unknown key type")
	}
	return
}

// ExportPrivateKeyAsPemStr exports the private key
func ExportPrivateKeyAsPemStr(sk interface{}) string {
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

var privateKey interface{}
var publicKey interface{}
var jwks = jwk.NewSet()

// GetPrivateKey returns the private key
func GetPrivateKey() interface{} {
	return privateKey
}

// GetPublicKey returns the public key
func GetPublicKey() interface{} {
	return publicKey
}

// GetJWKS returns the jwks
func GetJWKS() jwk.Set {
	return jwks
}

// LoadKey loads the private and public key
func LoadKey() {
	keyFileContent, err := ioutil.ReadFile(config.Get().Signing.KeyFile)
	if err != nil {
		panic(err)
	}
	switch config.Get().Signing.Alg {
	case oidc.RS256, oidc.RS384, oidc.RS512, oidc.PS256, oidc.PS384, oidc.PS512:
		sk, err := jwt.ParseRSAPrivateKeyFromPEM(keyFileContent)
		if err != nil {
			panic(err)
		}
		privateKey = sk
		publicKey = &sk.PublicKey
	case oidc.ES256, oidc.ES384, oidc.ES512:
		sk, err := jwt.ParseECPrivateKeyFromPEM(keyFileContent)
		if err != nil {
			panic(err)
		}
		privateKey = sk
		publicKey = &sk.PublicKey
	default:
		panic(fmt.Errorf("unknown signing alg"))
	}
	key, err := jwk.New(publicKey)
	if err != nil {
		panic(err)
	}
	if err = jwk.AssignKeyID(key); err != nil {
		panic(err)
	}
	if err = key.Set(jwk.KeyUsageKey, string(jwk.ForSignature)); err != nil {
		panic(err)
	}
	jwks.Add(key)
}
