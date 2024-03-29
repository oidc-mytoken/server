package revoke

import (
	"github.com/oidc-mytoken/utils/httpclient"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/oidcreqres"
)

// RefreshToken revokes a refresh token
func RefreshToken(rlog log.Ext1FieldLogger, provider model.Provider, rt string) *model.Response {
	if provider.Endpoints().Revocation == "" {
		return nil
	}
	req := oidcreqres.NewRTRevokeRequest(rt)
	httpRes, err := provider.AddClientAuthentication(httpclient.Do().R(), provider.Endpoints().Revocation).
		SetFormData(req.ToFormData()).
		SetError(&oidcreqres.OIDCErrorResponse{}).
		Post(provider.Endpoints().Revocation)
	if err != nil {
		rlog.WithError(err).Error()
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if errRes, ok := httpRes.Error().(*oidcreqres.OIDCErrorResponse); ok && errRes != nil && errRes.Error != "" {
		return &model.Response{
			Status:   httpRes.RawResponse.StatusCode,
			Response: model.OIDCError(errRes.Error, errRes.ErrorDescription),
		}
	}
	return nil
}
