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

	"github.com/go-oidfed/lib/jwx"
	"github.com/go-oidfed/lib/jwx/keymanagement/kms"
	"github.com/go-oidfed/lib/jwx/keymanagement/public"
	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/v4/jwa"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/config"
)

// GenerateMytokenSigningKeyPair generates a cryptographic key pair for mytoken signing with the algorithm specified in
// the mytoken config.
func GenerateMytokenSigningKeyPair() (sk crypto.Signer, pk crypto.PublicKey, err error) {
	return generateKeyPair(config.Get().Signing.Mytoken.Alg.SignatureAlgorithm, config.Get().Signing.Mytoken.RSAKeyLen)
}

// GenerateFederationSigningKeyPair generates a cryptographic key pair for federation signing with the algorithm
// specified in the config.
func GenerateFederationSigningKeyPair() (sk crypto.Signer, pk crypto.PublicKey, err error) {
	return generateKeyPair(
		config.Get().Features.Federation.Signing.Alg.SignatureAlgorithm,
		config.Get().Features.Federation.Signing.RSAKeyLen,
	)
}

// generateKeyPair generates a cryptographic key pair with the passed properties
func generateKeyPair(alg jwa.SignatureAlgorithm, rsaKeyLen int) (
	sk crypto.Signer, pk crypto.PublicKey,
	err error,
) {
	switch alg {
	case jwa.RS256(), jwa.RS384(), jwa.RS512(), jwa.PS256(), jwa.PS384(), jwa.PS512():
		if rsaKeyLen <= 0 {
			return nil, nil, errors.Errorf("%s specified, but no valid RSA key len", alg)
		}
		sk, err = rsa.GenerateKey(rand.Reader, rsaKeyLen)
	case jwa.ES256():
		sk, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case jwa.ES384():
		sk, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case jwa.ES512():
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
	KeyUsageFederation     = KeyUsage("oidfed")
	KeyUsageOIDCSigning    = KeyUsage("oidc comm")
)

type signingKeys map[KeyUsage]signingKeyMaterial

type signingKeyMaterial struct {
	SK   crypto.Signer
	PK   crypto.PublicKey
	JWKS jwx.JWKS
}

var keys signingKeys

// OIDC multi-key signing support
var (
	oidcKMS kms.BasicKeyManagementSystem
	oidcPKS public.PublicKeyStorage
)

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
func GetJWKS(usage KeyUsage) (jwks jwx.JWKS) {
	// OIDC signing uses multi-key KMS
	if usage == KeyUsageOIDCSigning {
		if oidcKMS == nil {
			return
		}
		vs := kms.KMSToVersatileSignerWithPKStorage(oidcKMS, oidcPKS)
		jwks, _ = vs.JWKS()
		return
	}

	// Other usages use single-key mode
	k, ok := keys[usage]
	if ok {
		jwks = k.JWKS
	}
	return
}

// GetVersatileSigner returns a VersatileSigner for the given key usage
func GetVersatileSigner(usage KeyUsage) jwx.VersatileSigner {
	// OIDC signing uses multi-key KMS
	if usage == KeyUsageOIDCSigning {
		if oidcKMS == nil {
			return nil
		}
		return kms.KMSToVersatileSignerWithPKStorage(oidcKMS, oidcPKS)
	}

	// Other usages use single-key mode
	k, ok := keys[usage]
	if !ok {
		return nil
	}
	var alg jwa.SignatureAlgorithm
	switch usage {
	case KeyUsageMytokenSigning:
		alg = config.Get().Signing.Mytoken.Alg.SignatureAlgorithm
	case KeyUsageFederation:
		alg = config.Get().Features.Federation.Signing.Alg.SignatureAlgorithm
	}
	return jwx.NewSingleKeyVersatileSigner(k.SK, alg)
}

// LoadMytokenSigningKey loads the private and public key for signing mytokens
func LoadMytokenSigningKey() {
	loadKey(
		config.Get().Signing.Mytoken.KeyFile, KeyUsageMytokenSigning,
		config.Get().Signing.Mytoken.Alg.SignatureAlgorithm,
	)
}

// LoadOIDCSigningKey loads the private and public key(s) for signing operations within OIDC communication
func LoadOIDCSigningKey() error {
	conf := config.Get().Signing.OIDC

	// Parse algorithms
	var algs []jwa.SignatureAlgorithm
	for _, algStr := range conf.Algorithms {
		alg, _ := jwa.LookupSignatureAlgorithm(algStr)
		algs = append(algs, alg)
	}

	// Default algorithm
	defaultAlg, _ := jwa.LookupSignatureAlgorithm(conf.DefaultAlgorithm)
	if defaultAlg.String() == "" && len(algs) > 0 {
		defaultAlg = algs[0]
	}

	// Create and load KMS
	kmsInst, err := kms.NewFilesystemKMSAndPublicKeyStorage(
		kms.FilesystemKMSConfig{
			KMSConfig: kms.KMSConfig{
				GenerateKeys: conf.GenerateKeys,
				Algs:         algs,
				DefaultAlg:   defaultAlg,
				RSAKeyLen:    conf.RSAKeyLen,
			},
			Dir:    conf.KeyDir,
			TypeID: "oidc",
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to create OIDC KMS")
	}
	if err = kmsInst.Load(); err != nil {
		return errors.Wrap(err, "failed to load OIDC signing keys")
	}
	oidcKMS = kmsInst
	if fsKMS, ok := kmsInst.(*kms.FilesystemKMS); ok {
		oidcPKS = fsKMS.PKs
	} else {
		return errors.New("failed to obtain OIDC public key storage from KMS")
	}
	return nil
}

// LoadFederationKey loads the private and public key for signing federation statements
func LoadFederationKey() {
	loadKey(
		config.Get().Features.Federation.Signing.KeyFile, KeyUsageFederation,
		config.Get().Features.Federation.Signing.Alg.SignatureAlgorithm,
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
	case jwa.RS256(), jwa.RS384(), jwa.RS512(), jwa.PS256(), jwa.PS384(), jwa.PS512():
		sk, err = jwt.ParseRSAPrivateKeyFromPEM(keyFileContent)
		if err != nil {
			panic(err)
		}
	case jwa.ES256(), jwa.ES384(), jwa.ES512():
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
	keyData.JWKS, _ = jwx.KeyToJWKS(keyData.PK, alg)
	keys[usage] = keyData
}
