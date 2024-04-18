package actions

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/unixtime"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/actionrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/actions/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/server/routes"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// HandleActions is the main entry function to handle the different actions of the action endpoint
func HandleActions(ctx *fiber.Ctx) error {
	actionInfo := pkg.CtxGetActionInfo(ctx)
	switch actionInfo.Action {
	case pkg.ActionRecreate:
		return handleRecreate(ctx, actionInfo.Code)
	case pkg.ActionVerifyEmail:
		return handleVerifyEmail(ctx, actionInfo.Code)
	case pkg.ActionRemoveFromCalendar:
		return handleRemoveFromCalendar(ctx, actionInfo.Code)
	case pkg.ActionUnsubscribeScheduled:
		return handleUnsubscribeScheduled(ctx, actionInfo.Code)
	}
	return ctxutils.RenderErrorPage(
		ctx, fiber.StatusBadRequest, model.BadRequestError("unknown action").
			CombinedMessage(),
	)
}

func handleRecreate(ctx *fiber.Ctx, code string) (err error) {
	rlog := logger.GetRequestLogger(ctx)
	var found bool
	var baseRequest string
	err = db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			var data actionrepo.RecreateData
			data, found, err = actionrepo.GetRecreateData(rlog, tx, code)
			if err != nil || !found {
				return err
			}
			var req api.GeneralMytokenRequest
			req.Issuer = data.Issuer
			if data.Name.Valid {
				req.Name = data.Name.String
			}
			req.Rotation = data.Rotation
			req.Capabilities = data.Capabilities
			if data.Restrictions != nil {
				restr := make(api.Restrictions, len(data.Restrictions))
				created := data.Created
				now := unixtime.Now()
				diff := now - created
				for i, r := range data.Restrictions {
					apiR := r.Restriction
					if r.NotBefore != 0 {
						apiR.NotBefore = int64(r.NotBefore + diff)
					}
					if r.ExpiresAt != 0 {
						apiR.ExpiresAt = int64(r.ExpiresAt + diff)
					}
					restr[i] = &apiR
				}
				req.Restrictions = restr
			}
			j, err := json.Marshal(req)
			if err != nil {
				return err
			}
			baseRequest = base64.RawURLEncoding.EncodeToString(j)
			return nil
		},
	)
	if err != nil {
		return ctxutils.RenderInternalServerErrorPage(ctx, err)
	}
	if !found {
		return ctxutils.RenderErrorPage(ctx, fiber.StatusNotFound, "recreation code not found")
	}
	return ctx.Redirect(fmt.Sprintf("/home?r=%s#mt", baseRequest), fiber.StatusSeeOther)
}

func handleVerifyEmail(ctx *fiber.Ctx, code string) error {
	rlog := logger.GetRequestLogger(ctx)
	verified, err := actionrepo.VerifyMail(rlog, nil, code)
	if err != nil {
		return ctxutils.RenderInternalServerErrorPage(ctx, err)
	}
	if !verified {
		return ctxutils.RenderErrorPage(ctx, http.StatusBadRequest, "code not valid or expired")
	}
	return ctxutils.RenderErrorPage(
		ctx, http.StatusOK, "The email address was successfully verified.", "Email Verified",
	)
}

func handleRemoveFromCalendar(ctx *fiber.Ctx, code string) error {
	rlog := logger.GetRequestLogger(ctx)
	err := actionrepo.UseRemoveCalendarCode(rlog, nil, code)
	if err != nil {
		return ctxutils.RenderInternalServerErrorPage(ctx, err)
	}
	return ctxutils.RenderErrorPage(
		ctx, http.StatusOK, "The token was successfully removed from the calendar.", "Token Removed from Calendar",
	)
}

func handleUnsubscribeScheduled(ctx *fiber.Ctx, code string) error {
	rlog := logger.GetRequestLogger(ctx)
	err := actionrepo.UseUnsubscribeFurtherNotificationsCode(rlog, nil, code)
	if err != nil {
		return ctxutils.RenderInternalServerErrorPage(ctx, err)
	}
	return ctxutils.RenderErrorPage(
		ctx, http.StatusOK, "You have successfully unsubscribed from further notifications of this kind.",
		"Unsubscribed",
	)
}

// CreateVerifyEmail creates an action url for verifying a mail address
func CreateVerifyEmail(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID) (string, error) {
	code := pkg.ActionInfo{
		Action: pkg.ActionVerifyEmail,
		Code:   pkg.NewCode(),
	}
	if err := actionrepo.AddVerifyEmailCode(rlog, tx, mtID, code.Code, pkg.CodeLifetimes[code.Action]); err != nil {
		return "", err
	}
	return routes.ActionsURL(code), nil
}

// CreateRecreateToken creates an action url for recreating a mytoken
func CreateRecreateToken(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID) (string, error) {
	code := pkg.ActionInfo{
		Action: pkg.ActionRecreate,
		Code:   pkg.NewCode(),
	}
	if err := actionrepo.AddRecreateTokenCode(rlog, tx, mtID, code.Code); err != nil {
		return "", err
	}
	return routes.ActionsURL(code), nil
}

// CreateRemoveFromCalendar creates an action url for removing a token from a calendar
func CreateRemoveFromCalendar(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, calendarName string) (
	string,
	error,
) {
	code := pkg.ActionInfo{
		Action: pkg.ActionRemoveFromCalendar,
		Code:   pkg.NewCode(),
	}
	if err := actionrepo.AddRemoveFromCalendarCode(rlog, tx, mtID, code.Code, calendarName); err != nil {
		return "", err
	}
	return routes.ActionsURL(code), nil
}

// GetUnsubscribeScheduled obtains the action code stored in the database for a scheduled notification and returns
// the action url
func GetUnsubscribeScheduled(rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, nid uint64) (
	string,
	error,
) {
	code, err := actionrepo.GetScheduledNotificationActionCode(rlog, tx, mtID, nid)
	if err != nil {
		return "", err
	}
	ac := pkg.ActionInfo{
		Action: pkg.ActionUnsubscribeScheduled,
		Code:   code,
	}
	return routes.ActionsURL(ac), nil
}
