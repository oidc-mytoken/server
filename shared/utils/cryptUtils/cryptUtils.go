package cryptUtils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"

	"github.com/oidc-mytoken/server/shared/utils"
)

const (
	saltLen       = 16
	keyIterations = 8
)

func deriveKey(password string, salt []byte, size int) ([]byte, []byte) {
	if salt == nil {
		salt = []byte(utils.RandASCIIString(saltLen))
	}
	return pbkdf2.Key([]byte(password), salt, keyIterations, size, sha512.New), salt
}

func AES128Encrypt(plain, password string) (string, error) {
	return aesEncrypt(plain, password, 16)
}
func AES192Encrypt(plain, password string) (string, error) {
	return aesEncrypt(plain, password, 24)
}
func AES256Encrypt(plain, password string) (string, error) {
	return aesEncrypt(plain, password, 32)
}

func AES128Decrypt(cipher, password string) (string, error) {
	return aesDecrypt(cipher, password, 16)
}
func AES192Decrypt(cipher, password string) (string, error) {
	return aesDecrypt(cipher, password, 24)
}
func AES256Decrypt(cipher, password string) (string, error) {
	return aesDecrypt(cipher, password, 32)
}

func aesEncrypt(plain, password string, keyLen int) (string, error) {
	key, salt := deriveKey(password, nil, keyLen)
	cipher, nonce, err := aesE(plain, key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%s", base64.StdEncoding.EncodeToString(salt), base64.StdEncoding.EncodeToString(nonce), base64.StdEncoding.EncodeToString(cipher)), nil
}

func aesDecrypt(cipher, password string, keyLen int) (string, error) {
	arr := strings.Split(cipher, "-")
	if len(arr) != 3 {
		return "", fmt.Errorf("malformed ciphertext")
	}
	salt, err := base64.StdEncoding.DecodeString(arr[0])
	if err != nil {
		return "", fmt.Errorf("malformed ciphertext: %s", err)
	}
	nonce, err := base64.StdEncoding.DecodeString(arr[1])
	if err != nil {
		return "", fmt.Errorf("malformed ciphertext: %s", err)
	}
	data, err := base64.StdEncoding.DecodeString(arr[2])
	if err != nil {
		return "", fmt.Errorf("malformed ciphertext: %s", err)
	}
	key, _ := deriveKey(password, salt, keyLen)
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
		//TODO
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
