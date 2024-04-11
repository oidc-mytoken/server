package cryptutils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLen       = 16
	keyIterations = 8
)

const malFormedCiphertext = "malformed ciphertext"

// RandomBytes returns size random bytes
func RandomBytes(size int) ([]byte, error) {
	r := make([]byte, size)
	_, err := rand.Read(r)
	return r, errors.WithStack(err)
}

func deriveKey(password string, salt []byte, size int) ([]byte, []byte, error) {
	if salt == nil {
		var err error
		salt, err = RandomBytes(saltLen)
		if err != nil {
			return nil, nil, err
		}
	}
	return pbkdf2.Key([]byte(password), salt, keyIterations, size, sha512.New), salt, nil
}

// AES128Encrypt encrypts a string using AES128 with the passed password
func AES128Encrypt(plain, password string) (string, error) {
	return aesEncrypt(plain, password, 16)
}

// AES192Encrypt encrypts a string using AES192 with the passed password
func AES192Encrypt(plain, password string) (string, error) {
	return aesEncrypt(plain, password, 24)
}

// AES256Encrypt encrypts a string using AES256 with the passed password
func AES256Encrypt(plain, password string) (string, error) {
	return aesEncrypt(plain, password, 32)
}

// AES128Decrypt decrypts a string using AES128 with the passed password
func AES128Decrypt(cipher, password string) (string, error) {
	return aesDecrypt(cipher, password, 16)
}

// AES192Decrypt decrypts a string using AES192 with the passed password
func AES192Decrypt(cipher, password string) (string, error) {
	return aesDecrypt(cipher, password, 24)
}

// AES256Decrypt decrypts a string using AES256 with the passed password
func AES256Decrypt(cipher, password string) (string, error) {
	return aesDecrypt(cipher, password, 32)
}

func aesEncrypt(plain, password string, keyLen int) (string, error) {
	key, salt, err := deriveKey(password, nil, keyLen)
	if err != nil {
		return "", err
	}
	ciph, err := AESEncrypt(plain, key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", base64.StdEncoding.EncodeToString(salt), ciph), nil
}

// AESEncrypt encrypts a string using the passed key
func AESEncrypt(plain string, key []byte) (string, error) {
	ciph, nonce, err := aesE(plain, key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", base64.StdEncoding.EncodeToString(nonce), base64.StdEncoding.EncodeToString(ciph)), nil
}

func aesDecrypt(cipher, password string, keyLen int) (string, error) {
	arr := strings.SplitN(cipher, "-", 2)
	salt, err := base64.StdEncoding.DecodeString(arr[0])
	if err != nil {
		return "", errors.Wrap(err, malFormedCiphertext)
	}
	key, _, err := deriveKey(password, salt, keyLen)
	if err != nil {
		return "", err
	}
	return AESDecrypt(arr[1], key)
}

// AESDecrypt decrypts a string using the passed key
func AESDecrypt(cipher string, key []byte) (string, error) {
	arr := strings.Split(cipher, "-")
	if len(arr) != 2 {
		return "", errors.New(malFormedCiphertext)
	}
	nonce, err := base64.StdEncoding.DecodeString(arr[0])
	if err != nil {
		return "", errors.Wrap(err, malFormedCiphertext)
	}
	data, err := base64.StdEncoding.DecodeString(arr[1])
	if err != nil {
		return "", errors.Wrap(err, malFormedCiphertext)
	}
	return aesD(data, nonce, key)
}

func aesE(plain string, key []byte) ([]byte, []byte, error) {
	gcm, err := createGCM(key)
	if err != nil {
		return nil, nil, err
	}
	nonce := []byte(utils.RandASCIIString(gcm.NonceSize()))
	ciph := gcm.Seal(nil, nonce, []byte(plain), nil)
	return ciph, nonce, nil
}

func aesD(cipher, nonce, key []byte) (string, error) {
	gcm, err := createGCM(key)
	if err != nil {
		return "", err
	}
	plain, err := gcm.Open(nil, nonce, cipher, nil)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(plain), nil
}

func createGCM(key []byte) (g cipher.AEAD, err error) {
	var block cipher.Block
	block, err = aes.NewCipher(key)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	g, err = cipher.NewGCM(block)
	err = errors.WithStack(err)
	return
}
