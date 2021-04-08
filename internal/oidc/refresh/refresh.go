package refresh

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/refreshtokenrepo"
	"github.com/oidc-mytoken/server/internal/oidc/oidcReqRes"
	"github.com/oidc-mytoken/server/shared/httpClient"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
)

// UpdateChangedRT is a function that should update a refresh token, it takes the old value as well as the new one
type UpdateChangedRT func(tokenID mtid.MTID, newRT, mytoken string) error

// Refresh uses an refresh token to obtain a new access token; if the refresh token changes, this is ignored
func Refresh(provider *config.ProviderConf, tokenID mtid.MTID, mytoken, rt, scopes, audiences string) (*oidcReqRes.OIDCTokenResponse, *oidcReqRes.OIDCErrorResponse, error) {
	return RefreshFlowAndUpdate(provider, tokenID, mytoken, rt, scopes, audiences, nil)
}

// RefreshFlowAndUpdate uses an refresh token to obtain a new access token; if the refresh token changes, the UpdateChangedRT function is used to update the refresh token
func RefreshFlowAndUpdate(provider *config.ProviderConf, tokenID mtid.MTID, mytoken, rt, scopes, audiences string, updateFnc UpdateChangedRT) (*oidcReqRes.OIDCTokenResponse, *oidcReqRes.OIDCErrorResponse, error) {
	req := oidcReqRes.NewRefreshRequest(rt, provider)
	req.Scopes = scopes
	req.Audiences = audiences
	httpRes, err := httpClient.Do().R().SetBasicAuth(provider.ClientID, provider.ClientSecret).SetFormData(req.ToFormData()).SetResult(&oidcReqRes.OIDCTokenResponse{}).SetError(&oidcReqRes.OIDCErrorResponse{}).Post(provider.Endpoints.Token)
	if err != nil {
		return nil, nil, err
	}
	if errRes, ok := httpRes.Error().(*oidcReqRes.OIDCErrorResponse); ok && errRes != nil && errRes.Error != "" {
		errRes.Status = httpRes.RawResponse.StatusCode
		return nil, errRes, nil
	}
	res, ok := httpRes.Result().(*oidcReqRes.OIDCTokenResponse)
	if !ok {
		return nil, nil, fmt.Errorf("could not unmarshal oidc response")
	}
	if res.RefreshToken != "" && res.RefreshToken != rt && updateFnc != nil {
		if err = updateFnc(tokenID, res.RefreshToken, mytoken); err != nil {
			log.WithError(err).Error()
			return res, nil, err
		}
	}
	return res, nil, nil
}

// RefreshFlowAndUpdateDB uses an refresh token to obtain a new access token; if the refresh token changes, it is updated in the database
func RefreshFlowAndUpdateDB(provider *config.ProviderConf, tokenID mtid.MTID, mytoken, rt, scopes, audiences string) (*oidcReqRes.OIDCTokenResponse, *oidcReqRes.OIDCErrorResponse, error) {
	return RefreshFlowAndUpdate(provider, tokenID, mytoken, rt, scopes, audiences, updateChangedRTInDB)
}

func updateChangedRTInDB(tokenID mtid.MTID, newRT, mytoken string) error {
	return refreshtokenrepo.UpdateRefreshToken(nil, tokenID, newRT, mytoken)
}
