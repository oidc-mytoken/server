package ssh

import (
	"fmt"
	"net"
	"time"

	"github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/sshrepo"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

func checkPubKey(ctx ssh.Context, key ssh.PublicKey) bool {
	sessionID := ctx.SessionID()
	rlog := logger.GetSSHRequestLogger(sessionID)
	sshUser := ctx.User()
	sshUserHash := hashUtils.SHA3_512Str([]byte(sshUser))
	sshKeyFP := gossh.FingerprintSHA256(key)
	ip := ctx.RemoteAddr().String()
	if addr, ok := ctx.RemoteAddr().(*net.TCPAddr); ok {
		ip = addr.IP.String()
	}
	userAgent := ctx.ClientVersion()
	log.WithFields(
		log.Fields{
			"user_hash":  sshUserHash,
			"key_fp":     sshKeyFP,
			"ip":         ip,
			"user_agent": userAgent,
		},
	).Debug("Check ssh pub key")

	info, err := sshrepo.GetSSHInfo(rlog, nil, sshKeyFP, sshUserHash)
	if err != nil {
		log.Errorf("%s", errorfmt.Full(err))
		return false
	}
	if !info.Enabled {
		log.WithField("user_hash", sshUserHash).WithField("key_fp", sshKeyFP).Trace("SSH grant not enabled")
		return false
	}
	mt, err := info.Decrypt(sshUser)
	if err != nil {
		log.WithField("user_hash", sshUserHash).WithField(
			"key_fp", sshKeyFP,
		).WithError(err).Error("Could not decrypt stored mytoken")
		return false
	}
	// We use tx=nil because we don't want to roll this back in case of other errors. If we come to this point the
	// ssh key was successfully used
	if err = sshrepo.UsedKey(rlog, nil, sshKeyFP, sshUserHash); err != nil {
		log.WithField("user_hash", sshUserHash).WithField(
			"key_fp", sshKeyFP,
		).WithError(err).Error("error while updating usage time")
		return false
	}
	ctx.SetValue("mytoken", mt)
	ctx.SetValue("ip", ip)
	ctx.SetValue("user_agent", userAgent)
	ctx.SetValue("session", sessionID)
	return true
}

// Serve starts the ssh server
func Serve() {
	ssh.Handle(handleSSHSession)

	log.WithField("port", 2222).Info("starting ssh server")
	fmt.Println("Starting ssh server on port 2222 ...")
	server := &ssh.Server{
		Addr:             ":2222",
		MaxTimeout:       30 * time.Second,
		IdleTimeout:      10 * time.Second,
		PublicKeyHandler: checkPubKey,
	}
	if err := server.SetOption(ssh.NoPty()); err != nil {
		log.WithError(err).Fatal()
	}
	log.WithError(server.ListenAndServe()).Fatal()
}
