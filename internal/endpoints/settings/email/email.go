package email

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/userrepo"
	"github.com/oidc-mytoken/server/internal/endpoints/actions"
	"github.com/oidc-mytoken/server/internal/endpoints/settings"
	my "github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/model"
	eventService "github.com/oidc-mytoken/server/internal/mytoken/event"
	"github.com/oidc-mytoken/server/internal/mytoken/event/pkg"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/rotation"
	"github.com/oidc-mytoken/server/internal/mytoken/universalmytoken"
	notifier "github.com/oidc-mytoken/server/internal/notifier/client"
	"github.com/oidc-mytoken/server/internal/notifier/server/mailing/mailtemplates"
	"github.com/oidc-mytoken/server/internal/utils/auth"
	"github.com/oidc-mytoken/server/internal/utils/cookies"
	"github.com/oidc-mytoken/server/internal/utils/ctxutils"
	"github.com/oidc-mytoken/server/internal/utils/logger"
)

// MailSettingsInfoResponse is a type for the response for listing mail settings
type MailSettingsInfoResponse struct {
	api.MailSettingsInfoResponse
	TokenUpdate *my.MytokenResponse `json:"token_update,omitempty"`
}

// SetTokenUpdate implements the pkg.TokenUpdatableResponse interface
func (res *MailSettingsInfoResponse) SetTokenUpdate(tokenUpdate *my.MytokenResponse) {
	res.TokenUpdate = tokenUpdate
}

// HandleGet handles GET requests to the email settings endpoint
func HandleGet(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle get email info request")
	var reqMytoken universalmytoken.UniversalMytoken

	return settings.HandleSettingsHelper(
		ctx, &reqMytoken, api.CapabilityEmailRead, &api.EventEmailSettingsListed, "", fiber.StatusOK,
		func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *model.Response) {
			info, err := userrepo.GetMail(rlog, tx, mt.ID)
			if err != nil {
				return nil, model.ErrorToInternalServerErrorResponse(err)
			}
			return &MailSettingsInfoResponse{
				MailSettingsInfoResponse: api.MailSettingsInfoResponse{
					EmailAddress:   info.Mail,
					EmailVerified:  info.MailVerified,
					PreferHTMLMail: info.PreferHTMLMail,
				},
			}, nil
		}, false,
	)
}

func HandlePut(ctx *fiber.Ctx) error {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle update email settings request")
	var req api.UpdateMailSettingsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err).Send(ctx)
	}
	if req.PreferHTMLMail == nil && req.EmailAddress == "" {
		return model.Response{
			Status:   fiber.StatusBadRequest,
			Response: model.BadRequestError("no request parameter given"),
		}.Send(ctx)
	}
	var reqMytoken universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &reqMytoken, ctx)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt, ctx.IP(), api.CapabilityEmail,
	)
	if errRes != nil {
		return errRes.Send(ctx)
	}
	var tokenUpdate *my.MytokenResponse
	clientMetaData := ctxutils.ClientMetaData(ctx)
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if req.PreferHTMLMail != nil {
				if err := userrepo.ChangePreferredMailType(rlog, tx, mt.ID, *req.PreferHTMLMail); err != nil {
					return err
				}
				eventComment := "to plain text"
				if *req.PreferHTMLMail {
					eventComment = "to html"
				}
				if err := eventService.LogEvent(
					rlog, tx, pkg.MTEvent{
						Event:          api.EventEmailMimetypeChanged,
						Comment:        eventComment,
						MTID:           mt.ID,
						ClientMetaData: *clientMetaData,
					},
				); err != nil {
					return err
				}
			}
			if req.EmailAddress != "" {
				if err := userrepo.ChangeEmail(rlog, tx, mt.ID, req.EmailAddress); err != nil {
					return err
				}
				verificationURL, err := actions.CreateVerifyEmail(rlog, tx, mt.ID)
				if err != nil {
					return err
				}
				mailInfo, err := userrepo.GetMail(rlog, tx, mt.ID)
				if err != nil {
					return err
				}
				if err = eventService.LogEvent(
					rlog, tx, pkg.MTEvent{
						Event:          api.EventEmailChanged,
						MTID:           mt.ID,
						Comment:        req.EmailAddress,
						ClientMetaData: *clientMetaData,
					},
				); err != nil {
					return err
				}
				notifier.SendTemplateEmail(
					req.EmailAddress, mailtemplates.SubjectVerifyMail, mailInfo.PreferHTMLMail,
					mailtemplates.TemplateVerifyMail, map[string]any{
						"issuer": config.Get().IssuerURL,
						"link":   verificationURL,
					},
				)
			}
			if err := usedRestriction.UsedOther(rlog, tx, mt.ID); err != nil {
				return err
			}

			tu, err := rotation.RotateMytokenAfterOtherForResponse(
				rlog, tx, reqMytoken.JWT, mt, *clientMetaData, reqMytoken.OriginalTokenType,
			)
			tokenUpdate = tu
			return err
		},
	); err != nil {
		return model.ErrorToInternalServerErrorResponse(err).Send(ctx)
	}
	if tokenUpdate != nil {
		return model.Response{
			Status:   fiber.StatusOK,
			Response: my.OnlyTokenUpdateRes{TokenUpdate: tokenUpdate},
			Cookies:  []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)},
		}.Send(ctx)
	}
	return model.Response{Status: fiber.StatusNoContent}.Send(ctx)
}
