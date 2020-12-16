package hashUtils

import (
	"crypto/sha512"
	"fmt"
)

// SHA512 hashes the passed data with sha512
func SHA512(data []byte) ([]byte, error) {
	hash, err := SHA512Str(data)
	return []byte(hash), err
}

// SHA512Str hashes the passed data with sha512
func SHA512Str(data []byte) (string, error) {
	sha := sha512.New()
	if _, err := sha.Write(data); err != nil {
		return "", err
	}
	hash := sha.Sum(nil)
	hashStr := fmt.Sprintf("%x", hash)
	return hashStr, nil
}
