package jws

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"

	"github.com/zachmann/mytoken/internal/config"
)

// GenerateRSAKeyPair generates an RSA key pair
func GenerateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey) {
	sk, _ := rsa.GenerateKey(rand.Reader, 1024)
	return sk, &sk.PublicKey
}

// ExportRSAPrivateKeyAsPemStr exports the private key
func ExportRSAPrivateKeyAsPemStr(privkey *rsa.PrivateKey) string {
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
func ParseRSAPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

var privateKey *rsa.PrivateKey
var publicKey *rsa.PublicKey

// GetPrivateKey returns the private key
func GetPrivateKey() *rsa.PrivateKey {
	return privateKey
}

// GetPublicKey returns the public key
func GetPublicKey() *rsa.PublicKey {
	return publicKey
}

// Init does init
func Init() {
	keyFileContent, err := ioutil.ReadFile(config.Get().Signing.KeyFile)
	if err != nil {
		panic(err)
	}
	sk, err := ParseRSAPrivateKeyFromPemStr(string(keyFileContent))
	if err != nil {
		panic(err)
	}
	privateKey = sk
	publicKey = &sk.PublicKey
}
