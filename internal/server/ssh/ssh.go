package ssh

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/model"
)

func decodeData(data, dataType string) ([]byte, error) {
	if data == "" {
		return nil, nil
	}
	switch strings.ToLower(dataType) {
	case api.SSHMimetypeJson:
		return []byte(data), nil
	case api.SSHMimetypeJsonBase64:
		d, err := base64.StdEncoding.DecodeString(data)
		return d, errors.WithStack(err)
	default:
		return nil, errors.New("unknown mimetype")
	}
}

func handleSSHSession(s ssh.Session) {
	err := _handleSSHSession(s)
	if err != nil {
		if err = writeError(s, err); err != nil {
			log.WithError(err).Error()
		}
	}
}

func writeString(s ssh.Session, str string) error {
	_, err := s.Write([]byte(str + "\n"))
	return err
}

func writeJSON(s ssh.Session, o interface{}) error {
	data, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return writeString(s, string(data))
}

func writeError(s ssh.Session, err error) error {
	return writeString(s, err.Error())
}

func writeErrRes(s ssh.Session, errRes *model.Response) error {
	return writeString(s, errRes.Response.(api.Error).CombinedMessage())
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
	case api.SSHRequestMytoken:
		return handleSSHMytoken(req, s)
	case api.SSHRequestAccessToken:
		return handleSSHAT(req, s)
	case api.SSHRequestTokenInfoIntrospect:
		return handleIntrospect(s)
	case api.SSHRequestTokenInfoHistory:
		return handleHistory(s)
	case api.SSHRequestTokenInfoSubtokens:
		return handleSubtokens(s)
	case api.SSHRequestTokenInfoListMytokens:
		return handleListMytokens(s)
	default:
		return errors.New("unknown request")
	}
}
