package settings

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	serverModel "github.com/oidc-mytoken/server/internal/model"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/server/paths"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
	"github.com/oidc-mytoken/server/internal/utils/mytokenutils"
)

// InitSettings initializes the settings metadata
func InitSettings() {
	apiPaths := paths.GetCurrentAPIPaths()
	settingsMetadata.GrantTypeEndpoint = utils.CombineURLPath(
		config.Get().IssuerURL, apiPaths.UserSettingEndpoint, "grants",
	)
}

var settingsMetadata = api.SettingsMetaData{
	GrantTypeEndpoint: "grants",
}

// HandleSettings handles Metadata requests to the settings endpoint
func HandleSettings(*fiber.Ctx) *serverModel.Response {
	return &serverModel.Response{
		Status:   fiber.StatusOK,
		Response: settingsMetadata,
	}
}

// HandleSettingsHelper is a helper wrapper function that handles various settings request with the help of a callback
func HandleSettingsHelper(
	ctx *fiber.Ctx,
	tx *sqlx.Tx,
	reqMytoken *universalmytoken.UniversalMytoken,
	requiredCapability api.Capability,
	logEvent *api.Event,
	eventComment string,
	okStatus int,
	callback func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *serverModel.Response),
	tokenGoneAfterCallback bool,
) *serverModel.Response {
	rlog := logger.GetRequestLogger(ctx)

	var res *serverModel.Response
	if err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			mt, errRes := auth.RequireValidMytoken(rlog, tx, reqMytoken, ctx)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
				rlog, tx, mt, ctxutils.ClientMetaData(ctx), requiredCapability,
			)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			rsp, errRes := callback(tx, mt)
			if errRes != nil {
				res = errRes
				return errors.New("rollback")
			}
			res = &serverModel.Response{
				Status:   fiber.StatusOK,
				Response: rsp,
			}
			if tokenGoneAfterCallback {
				return nil
			}
			var rollback bool
			res, rollback = mytokenutils.DoAfterRequestThingsOther(
				rlog, tx, res, mt, *ctxutils.ClientMetaData(ctx),
				*logEvent, eventComment, usedRestriction, reqMytoken.JWT, reqMytoken.OriginalTokenType,
			)
			if rollback {
				return errors.New("rollback")
			}
			return nil
		},
	); err != nil && res == nil {
		rlog.Errorf("%s", errorfmt.Full(err))
		res = serverModel.ErrorToInternalServerErrorResponse(err)
	}
	return res
}
