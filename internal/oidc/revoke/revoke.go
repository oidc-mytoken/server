package revoke

import (
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/oidcReqRes"
	"github.com/oidc-mytoken/server/shared/httpClient"
	pkgModel "github.com/oidc-mytoken/server/shared/model"
)

// RefreshToken revokes a refresh token
func RefreshToken(provider *config.ProviderConf, rt string) *model.Response {
	if provider.Endpoints.Revocation == "" {
		return nil
	}
	req := oidcReqRes.NewRTRevokeRequest(rt)
	httpRes, err := httpClient.Do().R().
		SetBasicAuth(provider.ClientID, provider.ClientSecret).
		SetFormData(req.ToFormData()).
		SetError(&oidcReqRes.OIDCErrorResponse{}).
		Post(provider.Endpoints.Revocation)
	if err != nil {
		log.WithError(err).Error()
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if errRes, ok := httpRes.Error().(*oidcReqRes.OIDCErrorResponse); ok && errRes != nil && errRes.Error != "" {
		return &model.Response{
			Status:   httpRes.RawResponse.StatusCode,
			Response: pkgModel.OIDCError(errRes.Error, errRes.ErrorDescription),
		}
	}
	return nil
}
