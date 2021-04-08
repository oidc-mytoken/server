package hashUtils

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
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

// HMACSHA512Str creates a hmac using sha512
func HMACSHA512Str(data, secret []byte) string {
	h := hmac.New(sha512.New, secret)
	hmac := h.Sum(data)
	return base64.StdEncoding.EncodeToString(hmac)
}
