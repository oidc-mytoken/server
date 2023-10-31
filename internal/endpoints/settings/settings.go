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
	event "github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
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
func HandleSettings(ctx *fiber.Ctx) error {
	res := serverModel.Response{
		Status:   fiber.StatusOK,
		Response: settingsMetadata,
	}
	return res.Send(ctx)
}

// HandleSettingsHelper is a helper wrapper function that handles various settings request with the help of a callback
func HandleSettingsHelper(
	ctx *fiber.Ctx,
	reqMytoken *universalmytoken.UniversalMytoken,
	requiredCapability api.Capability,
	logEvent *event.Event,
	okStatus int,
	callback func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *serverModel.Response),
	tokenGoneAfterCallback bool,
) error {
	rlog := logger.GetRequestLogger(ctx)
	mt, errRes := auth.RequireValidMytoken(rlog, nil, reqMytoken, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt, ctx.IP(), requiredCapability,
	)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	var tokenUpdate *my.MytokenResponse
	var rsp my.TokenUpdatableResponse
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) (err error) {
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
					rlog, tx, eventService.MTEvent{
						Event: logEvent,
						MTID:  mt.ID,
					}, *clientMetaData,
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
			return errRes.Send(ctx)
		}
		rlog.Errorf("%s", errorfmt.Full(err))
		return serverModel.ErrorToInternalServerErrorResponse(err).Send(ctx)
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
	return serverModel.Response{
		Status:   okStatus,
		Response: rsp,
		Cookies:  cake,
	}.Send(ctx)
}
