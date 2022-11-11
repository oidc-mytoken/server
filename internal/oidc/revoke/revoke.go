package revoke

import (
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/oidcreqres"
	"github.com/oidc-mytoken/server/shared/httpclient"
	pkgModel "github.com/oidc-mytoken/server/shared/model"
)

// RefreshToken revokes a refresh token
func RefreshToken(rlog log.Ext1FieldLogger, provider *config.ProviderConf, rt string) *model.Response {
	if provider.Endpoints.Revocation == "" {
		return nil
	}
	req := oidcreqres.NewRTRevokeRequest(rt)
	httpRes, err := httpclient.Do().R().
		SetBasicAuth(provider.ClientID, provider.ClientSecret).
		SetFormData(req.ToFormData()).
		SetError(&oidcreqres.OIDCErrorResponse{}).
		Post(provider.Endpoints.Revocation)
	if err != nil {
		rlog.WithError(err).Error()
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if errRes, ok := httpRes.Error().(*oidcreqres.OIDCErrorResponse); ok && errRes != nil && errRes.Error != "" {
		return &model.Response{
			Status:   httpRes.RawResponse.StatusCode,
			Response: pkgModel.OIDCError(errRes.Error, errRes.ErrorDescription),
		}
	}
	return nil
}
