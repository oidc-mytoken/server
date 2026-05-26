package ssh

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/internal/utils/mytokenutils"
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
		return nil, errors.New(
			fmt.Sprintf(
				"unknown mimetype; supported mimetypes are: %q",
				[]string{
					api.SSHMimetypeJson,
					api.SSHMimetypeJsonBase64,
				},
			),
		)
	}
}

func handleSSHSession(s ssh.Session) {
	defer func() {
		if r := recover(); r != nil {
			log.WithField("panic", r).Error("Panic in SSH session handler")
			_ = writeError(s, errors.New("Internal server error"))
		}
	}()
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
		return errors.New(fmt.Sprintf("Invalid Request\n%s", helpError))
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
	case api.SSHRequestTokenInfoNotifications:
		return handleTokenInfoNotifications(req, s)
	case api.SSHRequestRevoke:
		return handleSSHRevoke(req, s)
	case api.SSHRequestAddTag:
		return handleSSHAddTag(req, s)
	case api.SSHRequestRemoveTag:
		return handleSSHRemoveTag(req, s)
	case api.SSHRequestEmailGet:
		return handleSSHEmailGet(s)
	case api.SSHRequestEmailSet:
		return handleSSHEmailSet(req, s)
	case api.SSHRequestTagsList:
		return handleSSHTagsList(s)
	case api.SSHRequestTagCreate:
		return handleSSHTagCreate(req, s)
	case api.SSHRequestTagUpdate:
		return handleSSHTagUpdate(req, s)
	case api.SSHRequestTagDelete:
		return handleSSHTagDelete(req, s)
	case api.SSHRequestNotifications:
		return handleSSHNotificationsList(s)
	case api.SSHRequestNotificationCreate:
		return handleSSHNotificationCreate(req, s)
	case api.SSHRequestNotificationAddToken:
		return handleSSHNotificationAddToken(req, s)
	case api.SSHRequestNotificationRemoveToken:
		return handleSSHNotificationRemoveToken(req, s)
	case api.SSHRequestCalendars:
		return handleSSHCalendarsList(s)
	case api.SSHRequestCalendarCreate:
		return handleSSHCalendarCreate(req, s)
	case api.SSHRequestCalendarGet:
		return handleSSHCalendarGet(req, s)
	case api.SSHRequestCalendarUpdate:
		return handleSSHCalendarUpdate(req, s)
	case api.SSHRequestCalendarDelete:
		return handleSSHCalendarDelete(req, s)
	case api.SSHRequestCalendarAddMytoken:
		return handleSSHCalendarAddMytoken(req, s)
	case api.SSHRequestCalendarAddTag:
		return handleSSHCalendarAddTag(req, s)
	case api.SSHRequestCalendarRemoveTag:
		return handleSSHCalendarRemoveTag(req, s)
	case api.SSHRequestCalendarRemoveMytoken:
		return handleSSHCalendarRemoveMytoken(req, s)
	case api.SSHRequestNotificationUpdate:
		return handleSSHNotificationUpdate(req, s)
	case api.SSHRequestNotificationDelete:
		return handleSSHNotificationDelete(req, s)
	default:
		return errors.New(fmt.Sprintf("Unknown request\n%s", helpError))
	}
}

// sshSessionCtx holds the common context extracted from an SSH session
type sshSessionCtx struct {
	mt             *mytoken.Mytoken
	clientMetaData *api.ClientMetaData
	rlog           log.Ext1FieldLogger
}

// newSSHSessionCtx extracts common context from an SSH session
func newSSHSessionCtx(s ssh.Session) sshSessionCtx {
	ctx := s.Context()
	return sshSessionCtx{
		mt: ctx.Value("mytoken").(*mytoken.Mytoken),
		clientMetaData: &api.ClientMetaData{
			IP:        ctx.Value("ip").(string),
			UserAgent: ctx.Value("user_agent").(string),
		},
		rlog: logger.GetSSHRequestLogger(ctx.Value("session").(string)),
	}
}

// requireNotRevoked checks that the mytoken is not revoked
func (c sshSessionCtx) requireNotRevoked(s ssh.Session) error {
	errRes := auth.RequireMytokenNotRevoked(c.rlog, nil, c.mt, c.clientMetaData)
	if errRes != nil {
		return writeErrRes(s, errRes)
	}
	return nil
}

// requireCapability checks the mytoken's capability and restrictions
func (c sshSessionCtx) requireCapability(s ssh.Session, cap api.Capability) (*restrictions.Restriction, error) {
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		c.rlog, nil, c.mt, c.clientMetaData, cap,
	)
	if errRes != nil {
		return nil, writeErrRes(s, errRes)
	}
	return usedRestriction, nil
}

// doAfterRequest wraps the common post-request processing
func doAfterRequest(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, resIn *model.Response, mt *mytoken.Mytoken,
	clientMetaData api.ClientMetaData, event api.Event, eventComment string,
	usedRestriction *restrictions.Restriction, umt universalmytoken.UniversalMytoken,
) (*model.Response, error) {
	var rollback bool
	res, rollback := mytokenutils.DoAfterRequestThingsOther(
		rlog, tx, resIn, mt, clientMetaData,
		event, eventComment, usedRestriction, umt.JWT, umt.OriginalTokenType,
	)
	if rollback {
		return res, errors.New("rollback")
	}
	return res, nil
}

// transactWithResult runs a transaction with standard error handling
func transactWithResult(
	rlog log.Ext1FieldLogger,
	fn func(tx *sqlx.Tx) (*model.Response, error),
) (*model.Response, error) {
	var res *model.Response
	err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			var err error
			res, err = fn(tx)
			return err
		},
	)
	if err != nil && res == nil {
		return nil, err
	}
	return res, nil
}

// momModeResult holds the result of resolving MomID mode authentication
type momModeResult struct {
	id              mtid.MTID
	momMode         bool
	usedRestriction *restrictions.Restriction
}

// resolveMomMode handles MomID resolution and capability validation
func (c sshSessionCtx) resolveMomMode(
	s ssh.Session, reqMOMID string,
	capIfParent, capIfNotParent api.Capability,
) (*momModeResult, error) {
	momID := c.mt.ID.MomID()
	if reqMOMID != "" {
		momID = mtid.MOMID{MTID: mtid.FromHash(reqMOMID)}
	}
	id, momMode, errRes := auth.ValidateCapabilityWithMomMode(
		c.rlog, nil, capIfParent, capIfNotParent, c.mt, momID, c.clientMetaData,
	)
	if errRes != nil {
		return nil, writeErrRes(s, errRes)
	}
	usedRestriction, errRes := auth.RequireUsableRestrictionOther(c.rlog, nil, c.mt, c.clientMetaData)
	if errRes != nil {
		return nil, writeErrRes(s, errRes)
	}
	return &momModeResult{
		id:              id,
		momMode:         momMode,
		usedRestriction: usedRestriction,
	}, nil
}

const helpError = `Syntax for a request is:
	$ ssh <host_info> <action> [<mime_type> <data>]
See https://mytoken-docs.data.kit.edu/start/ssh/#using-the-ssh-grant for more information.
`
