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

	"github.com/dgrijalva/jwt-go"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/zachmann/mytoken/internal/config"
)

func GenerateKeyPair() (sk interface{}, pk interface{}, err error) {
	alg := config.Get().Signing.Alg
	switch alg {
	case oidc.RS256, oidc.RS384, oidc.RS512, oidc.PS256, oidc.PS384, oidc.PS512:
		keyLen := config.Get().Signing.RSAKeyLen
		if keyLen <= 0 {
			return nil, nil, fmt.Errorf("%s specified, but no valid RSA key len", alg)
		}
		sk, err = rsa.GenerateKey(rand.Reader, keyLen)
	case oidc.ES256:
		sk, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case oidc.ES384:
		sk, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case oidc.ES512:
		sk, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return nil, nil, fmt.Errorf("unknown signing algorithm '%s'", alg)
	}
	if err != nil {
		switch sk.(type) {
		case *rsa.PrivateKey:
			pk = &sk.(*rsa.PrivateKey).PublicKey
		case *ecdsa.PrivateKey:
			pk = &sk.(*ecdsa.PrivateKey).PublicKey
		default:
			err = fmt.Errorf("Something went wrong. We just created an unknown key type.")
		}
	}
	return
}

// ExportPrivateKeyAsPemStr exports the private key
func ExportPrivateKeyAsPemStr(sk interface{}) string {
	switch sk.(type) {
	case *rsa.PrivateKey:
		return exportRSAPrivateKeyAsPemStr(sk.(*rsa.PrivateKey))
	case *ecdsa.PrivateKey:
		return exportECPrivateKeyAsPemStr(sk.(*ecdsa.PrivateKey))
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

// ParseRSAPrivateKeyFromPemStr imports an private key
//func ParseRSAPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
//	block, _ := pem.Decode([]byte(privPEM))
//	if block == nil {
//		return nil, errors.New("failed to parse PEM block containing the key")
//	}
//
//	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
//	if err != nil {
//		return nil, err
//	}
//
//	return priv, nil
//}

var privateKey interface{}
var publicKey interface{}

// GetPrivateKey returns the private key
func GetPrivateKey() interface{} {
	return privateKey
}

// GetPublicKey returns the public key
func GetPublicKey() interface{} {
	return publicKey
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
}
