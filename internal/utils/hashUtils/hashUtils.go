package hashUtils

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"

	"golang.org/x/crypto/sha3"
)

// SHA512 hashes the passed data with sha512
func SHA512(data []byte) []byte {
	return []byte(SHA512Str(data))
}

// SHA512Str hashes the passed data with sha512
func SHA512Str(data []byte) string {
	hash := sha512.Sum512(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// HMACSHA3Str creates a hmac using sha3 512
func HMACSHA3Str(data, secret []byte) string {
	h := hmac.New(sha3.New512, secret)
	mac := h.Sum(data)
	return base64.StdEncoding.EncodeToString(mac)
}

// HMACBasedHash computes a hash-like value using HMAC
func HMACBasedHash(data []byte) string {
	return HMACSHA3Str(data, data)
}

// SHA3_256Str hashes the passed data with SHA3 256
//goland:noinspection GoSnakeCaseUsage
func SHA3_256Str(data []byte) string {
	hash := sha3.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// SHA3_512Str hashes the passed data with SHA3 512
//goland:noinspection GoSnakeCaseUsage
func SHA3_512Str(data []byte) string {
	hash := sha3.Sum512(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}
