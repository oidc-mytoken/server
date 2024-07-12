package refresh

import (
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/utils/httpclient"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/cryptstore"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/oidc/oidcreqres"
)

// UpdateChangedRT is a function that should update a refresh token, it takes the old value as well as the new one
type UpdateChangedRT func(rlog log.Ext1FieldLogger, tx *sqlx.Tx, tokenID mtid.MTID, newRT, mytoken string) error

// DoFlowWithoutUpdate uses a refresh token to obtain a new access token; if the refresh token changes, this is ignored
func DoFlowWithoutUpdate(
	rlog log.Ext1FieldLogger, provider model.Provider, tokenID mtid.MTID, mytoken, rt, scopes string,
	audiences []string,
) (*oidcreqres.OIDCTokenResponse, *oidcreqres.OIDCErrorResponse, error) {
	return DoFlowAndUpdate(rlog, nil, provider, tokenID, mytoken, rt, scopes, audiences, nil)
}

// DoFlowAndUpdate uses a refresh token to obtain a new access token; if the refresh token changes, the
// UpdateChangedRT function is used to update the refresh token
func DoFlowAndUpdate(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, provider model.Provider, tokenID mtid.MTID, mytoken, rt, scopes string,
	audiences []string,
	updateFnc UpdateChangedRT,
) (*oidcreqres.OIDCTokenResponse, *oidcreqres.OIDCErrorResponse, error) {
	req := oidcreqres.NewRefreshRequest(rt, provider.Audience())
	req.Scopes = scopes
	req.Audiences = audiences
	httpRes, err := provider.AddClientAuthentication(httpclient.Do().R(), provider.Endpoints().Token).
		SetFormDataFromValues(req.ToURLValues()).
		SetResult(&oidcreqres.OIDCTokenResponse{}).
		SetError(&oidcreqres.OIDCErrorResponse{}).
		Post(provider.Endpoints().Token)
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
		if err = updateFnc(rlog, tx, tokenID, res.RefreshToken, mytoken); err != nil {
			return res, nil, err
		}
	}
	return res, nil, nil
}

// DoFlowAndUpdateDB uses a refresh token to obtain a new access token; if the refresh token changes, it is
// updated in the database
func DoFlowAndUpdateDB(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, provider model.Provider, tokenID mtid.MTID, mytoken, rt, scopes string,
	audiences []string,
) (*oidcreqres.OIDCTokenResponse, *oidcreqres.OIDCErrorResponse, error) {
	return DoFlowAndUpdate(rlog, tx, provider, tokenID, mytoken, rt, scopes, audiences, updateChangedRTInDB)
}

func updateChangedRTInDB(rlog log.Ext1FieldLogger, tx *sqlx.Tx, tokenID mtid.MTID, newRT, mytoken string) error {
	return cryptstore.UpdateRefreshToken(rlog, tx, tokenID, newRT, mytoken)
}
