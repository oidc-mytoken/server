package refresh

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/cryptstore"
	"github.com/oidc-mytoken/server/internal/oidc/oidcreqres"
	"github.com/oidc-mytoken/server/shared/httpclient"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
)

// UpdateChangedRT is a function that should update a refresh token, it takes the old value as well as the new one
type UpdateChangedRT func(rlog log.Ext1FieldLogger, tokenID mtid.MTID, newRT, mytoken string) error

// DoFlowWithoutUpdate uses a refresh token to obtain a new access token; if the refresh token changes, this is ignored
func DoFlowWithoutUpdate(
	rlog log.Ext1FieldLogger, provider *config.ProviderConf, tokenID mtid.MTID, mytoken, rt, scopes, audiences string,
) (*oidcreqres.OIDCTokenResponse, *oidcreqres.OIDCErrorResponse, error) {
	return DoFlowAndUpdate(rlog, provider, tokenID, mytoken, rt, scopes, audiences, nil)
}

// DoFlowAndUpdate uses a refresh token to obtain a new access token; if the refresh token changes, the
// UpdateChangedRT function is used to update the refresh token
func DoFlowAndUpdate(
	rlog log.Ext1FieldLogger, provider *config.ProviderConf, tokenID mtid.MTID, mytoken, rt, scopes, audiences string,
	updateFnc UpdateChangedRT,
) (*oidcreqres.OIDCTokenResponse, *oidcreqres.OIDCErrorResponse, error) {
	req := oidcreqres.NewRefreshRequest(rt, provider)
	req.Scopes = scopes
	req.Audiences = audiences
	httpRes, err := httpclient.Do().R().
		SetBasicAuth(provider.ClientID, provider.ClientSecret).
		SetFormData(req.ToFormData()).
		SetResult(&oidcreqres.OIDCTokenResponse{}).
		SetError(&oidcreqres.OIDCErrorResponse{}).
		Post(provider.Endpoints.Token)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	if errRes, ok := httpRes.Error().(*oidcreqres.OIDCErrorResponse); ok && errRes != nil && errRes.Error != "" {
		errRes.Status = httpRes.RawResponse.StatusCode
		return nil, errRes, nil
	}
	res, ok := httpRes.Result().(*oidcreqres.OIDCTokenResponse)
	if !ok {
		return nil, nil, errors.New("could not unmarshal oidc response")
	}
	if res.RefreshToken != "" && res.RefreshToken != rt && updateFnc != nil {
		if err = updateFnc(rlog, tokenID, res.RefreshToken, mytoken); err != nil {
			return res, nil, err
		}
	}
	return res, nil, nil
}

// DoFlowAndUpdateDB uses a refresh token to obtain a new access token; if the refresh token changes, it is
// updated in the database
func DoFlowAndUpdateDB(
	rlog log.Ext1FieldLogger, provider *config.ProviderConf, tokenID mtid.MTID, mytoken, rt, scopes, audiences string,
) (*oidcreqres.OIDCTokenResponse, *oidcreqres.OIDCErrorResponse, error) {
	return DoFlowAndUpdate(rlog, provider, tokenID, mytoken, rt, scopes, audiences, updateChangedRTInDB)
}

func updateChangedRTInDB(rlog log.Ext1FieldLogger, tokenID mtid.MTID, newRT, mytoken string) error {
	return cryptstore.UpdateRefreshToken(rlog, nil, tokenID, newRT, mytoken)
}
