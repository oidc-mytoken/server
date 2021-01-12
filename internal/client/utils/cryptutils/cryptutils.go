package cryptutils

import (
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/zachmann/mytoken/internal/utils/cryptUtils"
)

// EncryptGPG encrypts the given string using the gpg key with the given id
func EncryptGPG(str, id string) (string, error) {
	cmd := exec.Command("/usr/bin/gpg", "-e", "--always-trust", "-r", id, "--armor")
	cmd.Stdin = strings.NewReader(str)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(out)
	return encoded, nil
}

// DecryptGPG decrypts the given ciphertext using the gpg key with the given id
func DecryptGPG(ciph, id string) (string, error) {
	cmd := exec.Command("/usr/bin/gpg", "-d", "-u", id)
	decoded, err := base64.StdEncoding.DecodeString(ciph)
	if err != nil {
		return "", err
	}
	cmd.Stdin = strings.NewReader(string(decoded))
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func EncryptPassword(str string) (string, error) {
	fmt.Printf("Enter encryption password: ")
	password, err := terminal.ReadPassword(syscall.Stdin)
	fmt.Println()
	if err != nil {
		return "", err
	}
	return cryptUtils.AES256Encrypt(str, string(password))
}

func DecryptPassword(ciph string) (string, error) {
	fmt.Printf("Enter decryption password: ")
	password, err := terminal.ReadPassword(syscall.Stdin)
	fmt.Println()
	if err != nil {
		return "", err
	}
	return cryptUtils.AES256Decrypt(ciph, string(password))
}
