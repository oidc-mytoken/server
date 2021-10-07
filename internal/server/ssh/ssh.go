package ssh

import (
	"encoding/base64"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func decodeData(data, dataType string) ([]byte, error) {
	if data == "" {
		return nil, nil
	}
	switch strings.ToLower(dataType) {
	case api.MIMETYPE_JSON:
		return []byte(data), nil
	case api.MIMETYPE_JSON_BASE64:
		d, err := base64.StdEncoding.DecodeString(data)
		return d, errors.WithStack(err)
	default:
		return nil, errors.New("unknown mimetype")
	}
}

func handleSSHSession(s ssh.Session) {
	err := _handleSSHSession(s)
	if err != nil {
		if _, err = s.Write([]byte(err.Error())); err != nil {
			log.WithError(err).Error()
		}
	}
}

func _handleSSHSession(s ssh.Session) (err error) {
	reqData := s.Command()
	noReqDataElements := len(reqData)
	if noReqDataElements != 1 && noReqDataElements != 3 {
		return errors.New("Invalid Request")
	}
	reqType := reqData[0]
	var req []byte = nil
	if noReqDataElements == 3 {
		req, err = decodeData(reqData[2], reqData[1])
		if err != nil {
			return
		}
	}

	switch reqType {
	case api.SSH_REQUEST_MYTOKEN:
		return handleSSHMytoken(req, s.Context())
	default:
		return errors.New("unknown request")
	}
}
