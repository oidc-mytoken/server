package settings

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	serverModel "github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	"github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/rotation"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	"github.com/oidc-mytoken/server/internal/server/paths"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/internal/utils/logger"
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
func HandleSettings(ctx *fiber.Ctx) *serverModel.Response {
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
	mt, errRes := auth.RequireValidMytoken(rlog, nil, reqMytoken, ctx)
	if errRes != nil {
		return errRes
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, tx, mt, ctxutils.ClientMetaData(ctx), requiredCapability,
	)
	if errRes != nil {
		return errRes
	}
	var tokenUpdate *my.MytokenResponse
	var rsp my.TokenUpdatableResponse
	if err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) (err error) {
			rsp, errRes = callback(tx, mt)
			if errRes != nil {
				return fmt.Errorf("dummy")
			}
			if tokenGoneAfterCallback {
				return
			}
			clientMetaData := ctxutils.ClientMetaData(ctx)
			if logEvent != nil {
				if err = eventService.LogEvent(
					rlog, tx, pkg.MTEvent{
						Event:          *logEvent,
						MTID:           mt.ID,
						ClientMetaData: *clientMetaData,
						Comment:        eventComment,
					},
				); err != nil {
					return
				}
			}
			if usedRestriction != nil {
				if err = usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
					return
				}
			}
			tokenUpdate, err = rotation.RotateMytokenAfterOtherForResponse(
				rlog, tx, reqMytoken.JWT, mt, *clientMetaData, reqMytoken.OriginalTokenType,
			)
			return
		},
	); err != nil {
		if errRes != nil {
			return errRes
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return serverModel.ErrorToInternalServerErrorResponse(err)
	}

	var cake []*fiber.Cookie
	if tokenUpdate != nil {
		if rsp == nil {
			rsp = &my.OnlyTokenUpdateRes{}
		}
		rsp.SetTokenUpdate(tokenUpdate)
		cake = []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)}
		okStatus = fiber.StatusOK
	}
	return &serverModel.Response{
		Status:   okStatus,
		Response: rsp,
		Cookies:  cake,
	}
}
