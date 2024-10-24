package email

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

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
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
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
func HandleGet(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle get email info request")
	var reqMytoken universalmytoken.UniversalMytoken

	return settings.HandleSettingsHelper(
		ctx, nil, &reqMytoken, api.CapabilityEmailRead, &api.EventEmailSettingsListed, "", fiber.StatusOK,
		func(tx *sqlx.Tx, mt *mytoken.Mytoken) (my.TokenUpdatableResponse, *model.Response) {
			info, err := userrepo.GetMail(rlog, tx, mt.ID)
			if err != nil {
				return nil, model.ErrorToInternalServerErrorResponse(err)
			}
			return &MailSettingsInfoResponse{
				MailSettingsInfoResponse: api.MailSettingsInfoResponse{
					EmailAddress:   info.Mail.String,
					EmailVerified:  info.MailVerified,
					PreferHTMLMail: info.PreferHTMLMail,
				},
			}, nil
		}, false,
	)
}

func changeEmailAddress(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, email string,
	clientMetaData *api.ClientMetaData,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if err := userrepo.ChangeEmail(rlog, tx, mtID, email); err != nil {
				return err
			}
			verificationURL, err := actions.CreateVerifyEmail(rlog, tx, mtID)
			if err != nil {
				return err
			}
			mailInfo, err := userrepo.GetMail(rlog, tx, mtID)
			if err != nil {
				return err
			}
			if err = eventService.LogEvent(
				rlog, tx, pkg.MTEvent{
					Event:          api.EventEmailChanged,
					MTID:           mtID,
					Comment:        email,
					ClientMetaData: *clientMetaData,
				},
			); err != nil {
				return err
			}
			notifier.SendTemplateEmail(
				email, mailtemplates.SubjectVerifyMail, mailInfo.PreferHTMLMail,
				mailtemplates.TemplateVerifyMail, map[string]any{
					"issuer": config.Get().IssuerURL,
					"link":   verificationURL,
				},
			)
			return nil
		},
	)
}

func changePreferredMimeType(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, mtID mtid.MTID, preferHTML bool,
	clientMetaData *api.ClientMetaData,
) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {

			if err := userrepo.ChangePreferredMailType(rlog, tx, mtID, preferHTML); err != nil {
				return err
			}
			eventComment := "to plain text"
			if preferHTML {
				eventComment = "to html"
			}
			if err := eventService.LogEvent(
				rlog, tx, pkg.MTEvent{
					Event:          api.EventEmailMimetypeChanged,
					Comment:        eventComment,
					MTID:           mtID,
					ClientMetaData: *clientMetaData,
				},
			); err != nil {
				return err
			}
			return nil
		},
	)
}

// HandlePut handles PUT requests to the email settings endpoint, i.e. it updates email settings
func HandlePut(ctx *fiber.Ctx) *model.Response {
	rlog := logger.GetRequestLogger(ctx)
	rlog.Debug("Handle update email settings request")
	var req api.UpdateMailSettingsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return model.ErrorToBadRequestErrorResponse(err)
	}
	if req.PreferHTMLMail == nil && req.EmailAddress == "" {
		return model.BadRequestErrorResponse("no request parameter given")
	}
	var reqMytoken universalmytoken.UniversalMytoken
	mt, errRes := auth.RequireValidMytoken(rlog, nil, &reqMytoken, ctx)
	if errRes != nil {
		return errRes
	}
	usedRestriction, errRes := auth.RequireCapabilityAndRestrictionOther(
		rlog, nil, mt, ctxutils.ClientMetaData(ctx), api.CapabilityEmail,
	)
	if errRes != nil {
		return errRes
	}
	var tokenUpdate *my.MytokenResponse
	clientMetaData := ctxutils.ClientMetaData(ctx)
	if err := db.Transact(
		rlog, func(tx *sqlx.Tx) error {
			if req.PreferHTMLMail != nil {
				if err := changePreferredMimeType(rlog, tx, mt.ID, *req.PreferHTMLMail, clientMetaData); err != nil {
					return err
				}
			}
			if req.EmailAddress != "" {
				if err := changeEmailAddress(rlog, tx, mt.ID, req.EmailAddress, clientMetaData); err != nil {
					return err
				}
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
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if tokenUpdate != nil {
		return &model.Response{
			Status:   fiber.StatusOK,
			Response: my.OnlyTokenUpdateRes{TokenUpdate: tokenUpdate},
			Cookies:  []*fiber.Cookie{cookies.MytokenCookie(tokenUpdate.Mytoken)},
		}
	}
	return &model.Response{Status: fiber.StatusNoContent}
}
