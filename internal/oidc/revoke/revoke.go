package revoke

import (
	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/oidc/oidcReqRes"
	pkgModel "github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/httpClient"
)

// RefreshToken revokes a refresh token
func RefreshToken(provider *config.ProviderConf, rt string) *model.Response {
	if len(provider.Endpoints.Revocation) == 0 {
		return nil
	}
	req := oidcReqRes.NewRTRevokeRequest(rt)
	httpRes, err := httpClient.Do().R().SetBasicAuth(provider.ClientID, provider.ClientSecret).SetFormData(req.ToFormData()).SetError(&oidcReqRes.OIDCErrorResponse{}).Post(provider.Endpoints.Revocation)
	if err != nil {
		return model.ErrorToInternalServerErrorResponse(err)
	}
	if errRes, ok := httpRes.Error().(*oidcReqRes.OIDCErrorResponse); ok && errRes != nil && len(errRes.Error) > 0 {
		return &model.Response{
			Status:   httpRes.RawResponse.StatusCode,
			Response: pkgModel.OIDCError(errRes.Error, errRes.ErrorDescription),
		}
	}
	return nil
}
