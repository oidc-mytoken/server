package hashUtils

import (
	"crypto/sha512"
	"fmt"
)

func SHA512(data []byte) ([]byte, error) {
	hash, err := SHA512Str(data)
	return []byte(hash), err
}
func SHA512Str(data []byte) (string, error) {
	sha := sha512.New()
	if _, err := sha.Write(data); err != nil {
		return "", err
	}
	hash := sha.Sum(nil)
	hashStr := fmt.Sprintf("%x", hash)
	return hashStr, nil
}
