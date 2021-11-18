package ssh

import (
	"fmt"

	"github.com/gliderlabs/ssh"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	gossh "golang.org/x/crypto/ssh"
)

func handleSSHMytoken(req []byte, s ssh.Session) error {
	// ctx := s.Context()
	fp256 := gossh.FingerprintSHA256(s.PublicKey())
	s.Write([]byte(fmt.Sprintf("256: %s\n", fp256)))
	return errors.New(api.ErrorNYI.CombinedMessage())
}
