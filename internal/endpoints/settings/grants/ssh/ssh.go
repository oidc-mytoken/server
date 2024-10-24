package ssh

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/grantrepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/transfercoderepo"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/sshrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/settings"
	request "github.com/oidc-mytoken/server/internal/endpoints/settings/grants/ssh/pkg"
	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/polling"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/model/profiled"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/oidc/authcode"
	"github.com/oidc-mytoken/server/internal/server/ssh"
	"github.com/oidc-mytoken/server/internal/utils/cryptutils"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/hashutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleGetSSHInfo handles requests to return information about a user's ssh keys
func HandleGetSSHInfo(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle get ssh info request")
	var reqMytoken universalmytoken.UniversalMytoken
	return settings.HandleSettingsHelper(
		ctx, nil, &reqMytoken, api.CapabilitySSHGrantRead, &api.EventSSHKeyListed, "", fiber.StatusOK,
		func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *model.Response) {
			info, err := sshrepo.GetAllSSHInfo(rlog, tx, mt.ID)
			_, err = db.ParseError(err)
			if err != nil {
				rlog.Errorf("%s", errorfmt.Full(err))
				return nil, model.ErrorToInternalServerErrorResponse(err)
			}
			grantEnabled, err := grantrepo.GrantEnabled(rlog, tx, mt.ID, model.GrantTypeSSH)
			if err != nil {
				rlog.Errorf("%s", errorfmt.Full(err))
				return nil, model.ErrorToInternalServerErrorResponse(err)
			}
			return &request.SSHInfoResponse{
				SSHInfoResponse: api.SSHInfoResponse{
					GrantEnabled: grantEnabled,
					SSHKeyInfo:   info,
				},
			}, nil
		}, false,
	)
}

// HandleDeleteSSHKey handles requests to delete a user's ssh public key
func HandleDeleteSSHKey(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle delete ssh key request")
	req := request.SSHKeyDeleteRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	rlog.Trace("Parsed delete ssh key request")
	if req.SSHKeyFingerprint == "" {
		if req.SSHKey == "" {
			return model.BadRequestErrorResponse(
				"One of the required parameters 'ssh_key' or 'ssh_key_hash' must be given",
			)
		}
		sshKey, _, _, _, err := gossh.ParseAuthorizedKey([]byte(req.SSHKey))
		if err != nil {
			return model.BadRequestErrorResponse("could not parse ssh public key")
		}
		req.SSHKeyFingerprint = gossh.FingerprintSHA256(sshKey)
	}

	return settings.HandleSettingsHelper(
		ctx, nil, &req.Mytoken, api.CapabilitySSHGrant, nil, "", fiber.StatusNoContent,
		func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *model.Response) {
			if err := sshrepo.Delete(rlog, tx, mt.ID, req.SSHKeyFingerprint); err != nil {
				rlog.Errorf("%s", errorfmt.Full(err))
				return nil, model.ErrorToInternalServerErrorResponse(err)
			}
			return nil, nil
		}, true,
	)
}

// HandlePost handles POST requests to the ssh grant endpoint, this includes the initial request to add an ssh public
// key as well as the following polling requests.
func HandlePost(ctx *fiber.Ctx) *model.Response {
	grantType, err := ctxutils.GetGrantType(ctx)
	if err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	switch grantType {
	case model.GrantTypeMytoken:
		return handleAddSSHKey(ctx)
	case model.GrantTypePollingCode:
		return handlePollingCode(ctx)
	default:
		return &model.Response{
			Status:   fiber.StatusBadRequest,
			Response: api.ErrorUnsupportedGrantType,
		}
	}
}

// handleAddSSHKey handles the initial request to add an ssh public key
func handleAddSSHKey(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle add ssh key request")
	req := request.SSHKeyAddRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	rlog.Trace("Parsed add ssh key request")
	if req.SSHKey == "" {
		return model.BadRequestErrorResponse("required parameter 'ssh_key' is missing")
	}
	sshKey, _, _, _, err := gossh.ParseAuthorizedKey([]byte(req.SSHKey))
	if err != nil {
		return model.BadRequestErrorResponse("could not parse ssh public key")
	}
	sshKeyFP := gossh.FingerprintSHA256(sshKey)
	if len(req.Capabilities) == 0 {
		req.Capabilities = api.Capabilities{api.CapabilityAT}
	}

	return settings.HandleSettingsHelper(
		ctx, nil, &req.Mytoken, api.CapabilitySSHGrant, &api.EventSSHKeyAdded, "", fiber.StatusOK,
		func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *model.Response) {
			return handleAddSSHSettingsCallback(rlog, ctx, &req, sshKeyFP, tx, mt)
		}, false,
	)
}

func handleAddSSHSettingsCallback(
	rlog log.Ext1FieldLogger, ctx *fiber.Ctx, req *request.SSHKeyAddRequest,
	sshKeyFP string,
	tx *sqlx.Tx,
	mt *mytoken.Mytoken,
) (
	my.TokenUpdatableResponse,
	*model.Response,
) {
	grantEnabled, err := grantrepo.GrantEnabled(rlog, tx, mt.ID, model.GrantTypeSSH)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return nil, model.ErrorToInternalServerErrorResponse(err)
	}
	if !grantEnabled {
		return nil, &model.Response{
			Status: fiber.StatusForbidden,
			Response: api.Error{
				Error:            api.ErrorStrAccessDenied,
				ErrorDescription: "'ssh' grant type not enabled",
			},
		}
	}
	allSSHKeys, err := sshrepo.GetAllSSHInfo(rlog, tx, mt.ID)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return nil, model.ErrorToInternalServerErrorResponse(err)
	}
	for _, k := range allSSHKeys {
		if k.SSHKeyFingerprint == sshKeyFP {
			return nil, &model.Response{
				Status: fiber.StatusConflict,
				Response: api.Error{
					Error:            api.ErrorStrInvalidRequest,
					ErrorDescription: "ssh key already added",
				},
			}
		}
	}

	mytokenReq := &my.AuthCodeFlowRequest{
		OIDCFlowRequest: my.OIDCFlowRequest{
			GeneralMytokenRequest: profiled.GeneralMytokenRequest{
				GeneralMytokenRequest: api.GeneralMytokenRequest{
					Issuer: mt.OIDCIssuer,
					Name:   embedSSHKeyNameInMTName(req.Name),
				},
				Restrictions: profiled.Restrictions{Restrictions: req.Restrictions},
				Capabilities: profiled.Capabilities{Capabilities: req.Capabilities},
				GrantType:    model.GrantTypeOIDCFlow,
				ResponseType: model.ResponseTypeToken,
			},
			OIDCFlowAttrs: my.OIDCFlowAttrs{
				OIDCFlow: model.OIDCFlowAuthorizationCode,
			},
		},
		AuthCodeFlowAttrs: my.AuthCodeFlowAttrs{
			ClientType: api.ClientTypeNative,
		},
	}
	res := authcode.StartAuthCodeFlow(ctx, mytokenReq)
	if res.Status >= 400 {
		return nil, res
	}
	authRes := res.Response.(api.AuthCodeFlowResponse)
	if err = transfercoderepo.LinkPollingCodeToSSHKey(rlog, tx, authRes.PollingCode, sshKeyFP); err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return nil, model.ErrorToInternalServerErrorResponse(err)
	}
	return &request.SSHKeyAddResponse{
		SSHKeyAddResponse: api.SSHKeyAddResponse{
			AuthCodeFlowResponse: authRes,
		},
	}, nil
}

const mtNamePrefix = "mytoken for ssh-grant"

func embedSSHKeyNameInMTName(keyName string) string {
	if keyName == "" {
		return mtNamePrefix
	}
	return mtNamePrefix + ": " + keyName
}

func extractSSHKeyNameFromMTName(mtName string) string {
	prefixLen := len(mtNamePrefix) + 2 // +2 because of ": "
	if len(mtName) <= prefixLen {
		return ""
	}
	return mtName[prefixLen:]
}

// handlePollingCode handles the polling requests to finish adding an ssh public key
func handlePollingCode(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	req := my.NewPollingCodeRequest()
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	clientMetaData := ctxutils.ClientMetaData(ctx)
	mt, token, status, errRes := polling.CheckPollingCodeReq(rlog, req, *clientMetaData, true)
	if errRes != nil {
		return errRes
	}
	user := utils.RandASCIIString(16)
	userHash := hashutils.SHA3_512Str([]byte(user))
	encryptedMT, err := cryptutils.AES256Encrypt(token, user)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	mtName, err := mytokenrepohelper.GetMTName(rlog, nil, mt.ID)
	if err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	name := extractSSHKeyNameFromMTName(mtName.String)
	data := sshrepo.SSHInfoIn{
		MTID:           mt.ID,
		Name:           db.NewNullString(name),
		KeyFingerprint: status.SSHKeyFingerprint.String,
		UserHash:       userHash,
		EncryptedMT:    encryptedMT,
	}
	if err = sshrepo.Insert(rlog, nil, data); err != nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		return model.ErrorToInternalServerErrorResponse(err)
	}
	return &model.Response{
		Status: fiber.StatusOK,
		Response: api.SSHKeyAddFinalResponse{
			SSHUser:       user,
			SSHHostConfig: ssh.CreateHostConfigEntry(user, name),
		},
	}
}
