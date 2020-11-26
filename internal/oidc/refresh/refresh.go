package refresh

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/httpClient"
	"github.com/zachmann/mytoken/internal/oidc/oidcReqRes"
)

type UpdateChangedRT func(oldRT, newRT string) error

func Refresh(provider *config.ProviderConf, rt string, scopes, audiences string) (*oidcReqRes.OIDCTokenResponse, *oidcReqRes.OIDCErrorResponse, error) {
	return RefreshFlowAndUpdate(provider, rt, scopes, audiences, nil)
}

func RefreshFlowAndUpdate(provider *config.ProviderConf, rt string, scopes, audiences string, updateFnc UpdateChangedRT) (*oidcReqRes.OIDCTokenResponse, *oidcReqRes.OIDCErrorResponse, error) {
	req := oidcReqRes.NewRefreshRequest(rt)
	req.Scopes = scopes
	req.Audiences = audiences
	httpRes, err := httpClient.Do().R().SetBasicAuth(provider.ClientID, provider.ClientSecret).SetFormData(req.ToFormData()).SetResult(&oidcReqRes.OIDCTokenResponse{}).SetError(&oidcReqRes.OIDCErrorResponse{}).Post(provider.Provider.Endpoint().TokenURL)
	if err != nil {
		return nil, nil, err
	}
	if errRes, ok := httpRes.Error().(*oidcReqRes.OIDCErrorResponse); ok && errRes != nil {
		errRes.Status = httpRes.RawResponse.StatusCode
		return nil, errRes, nil
	}
	res, ok := httpRes.Result().(*oidcReqRes.OIDCTokenResponse)
	if !ok {
		return nil, nil, fmt.Errorf("could not unmarshal oidc response")
	}
	if res.RefreshToken != "" && res.RefreshToken != rt && updateFnc != nil {
		if err := updateFnc(rt, res.RefreshToken); err != nil {
			log.WithError(err).Error()
			return res, nil, err
		}
	}
	return res, nil, nil
}

func RefreshFlowAndUpdateDB(provider *config.ProviderConf, rt string, scopes, audiences string) (*oidcReqRes.OIDCTokenResponse, *oidcReqRes.OIDCErrorResponse, error) {
	return RefreshFlowAndUpdate(provider, rt, scopes, audiences, updateChangedRTInDB)
}

func updateChangedRTInDB(oldRT, newRT string) error {
	_, err := db.DB().Exec(`UPDATE SuperTokens SET refresh_token=? WHERE refresh_token=?`, newRT, oldRT)
	return err
}