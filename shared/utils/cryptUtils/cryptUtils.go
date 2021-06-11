package cryptUtils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	mathRand "math/rand"
	"strings"

	"golang.org/x/crypto/pbkdf2"

	"github.com/oidc-mytoken/server/shared/utils"
)

const (
	saltLen       = 16
	keyIterations = 8
)

const malFormedCiphertext = "malformed ciphertext"
const malFormedCiphertextFmt = malFormedCiphertext + ": %s"

// RandomBytes returns size random bytes
func RandomBytes(size int) []byte {
	r := make([]byte, size)
	if _, err := rand.Read(r); err != nil {
		_, _ = mathRand.Read(r)
	}
	return r
}

func deriveKey(password string, salt []byte, size int) ([]byte, []byte) {
	if salt == nil {
		salt = RandomBytes(saltLen)
	}
	return pbkdf2.Key([]byte(password), salt, keyIterations, size, sha512.New), salt
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
	key, salt := deriveKey(password, nil, keyLen)
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
		return "", fmt.Errorf(malFormedCiphertextFmt, err)
	}
	key, _ := deriveKey(password, salt, keyLen)
	return AESDecrypt(arr[1], key)
}

// AESDecrypt decrypts a string using the passed key
func AESDecrypt(cipher string, key []byte) (string, error) {
	arr := strings.Split(cipher, "-")
	if len(arr) != 2 {
		return "", fmt.Errorf(malFormedCiphertext)
	}
	nonce, err := base64.StdEncoding.DecodeString(arr[0])
	if err != nil {
		return "", fmt.Errorf(malFormedCiphertextFmt, err)
	}
	data, err := base64.StdEncoding.DecodeString(arr[1])
	if err != nil {
		return "", fmt.Errorf(malFormedCiphertextFmt, err)
	}
	return aesD(data, nonce, key)
}

func aesE(plain string, key []byte) ([]byte, []byte, error) {
	gcm, err := createGCM(key)
	if err != nil {
		return nil, nil, err
	}
	nonce := []byte(utils.RandASCIIString(gcm.NonceSize()))
	cipher := gcm.Seal(nil, nonce, []byte(plain), nil)
	return cipher, nonce, nil
}

func aesD(cipher, nonce, key []byte) (string, error) {
	gcm, err := createGCM(key)
	if err != nil {
		return "", err
	}
	plain, err := gcm.Open(nil, nonce, cipher, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func createGCM(key []byte) (g cipher.AEAD, err error) {
	var block cipher.Block
	block, err = aes.NewCipher(key)
	if err != nil {
		return
	}
	g, err = cipher.NewGCM(block)
	return
}
