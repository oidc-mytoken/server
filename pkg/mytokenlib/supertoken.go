package mytokenlib

import (
	"errors"
	"fmt"
	"time"

	"github.com/oidc-mytoken/server/internal/endpoints/token/super/pkg"
	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/httpClient"
	"github.com/oidc-mytoken/server/shared/supertoken/capabilities"
	"github.com/oidc-mytoken/server/shared/supertoken/restrictions"
	"github.com/oidc-mytoken/server/shared/supertoken/token"
)

func (my *Mytoken) GetSuperToken(req interface{}) (string, error) {
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&pkg.SuperTokenResponse{}).SetError(&model.APIError{}).Post(my.SuperTokenEndpoint)
	if err != nil {
		return "", newMytokenErrorFromError("error while sending http request", err)
	}
	if e := resp.Error(); e != nil {
		if errRes := e.(*model.APIError); errRes != nil && errRes.Error != "" {
			return "", &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	stRes, ok := resp.Result().(*pkg.SuperTokenResponse)
	if !ok {
		return "", &MytokenError{
			err: "unexpected response from mytoken server",
		}
	}
	return stRes.SuperToken, nil
}

func (my *Mytoken) GetSuperTokenBySuperToken(superToken, issuer string, restrictions restrictions.Restrictions, capabilities, subtokenCapabilities capabilities.Capabilities, responseType model.ResponseType, name string) (string, error) {
	req := pkg.SuperTokenFromSuperTokenRequest{
		Issuer:               issuer,
		GrantType:            model.GrantTypeSuperToken,
		SuperToken:           token.Token(superToken),
		Restrictions:         restrictions,
		Capabilities:         capabilities,
		SubtokenCapabilities: subtokenCapabilities,
		Name:                 name,
		ResponseType:         responseType,
	}
	return my.GetSuperToken(req)
}

func (my *Mytoken) GetSuperTokenByTransferCode(transferCode string) (string, error) {
	req := pkg.ExchangeTransferCodeRequest{
		GrantType:    model.GrantTypeTransferCode,
		TransferCode: transferCode,
	}
	return my.GetSuperToken(req)
}

func (my *Mytoken) GetSuperTokenByAuthorizationFlow(issuer string, restrictions restrictions.Restrictions, capabilities, subtokenCapabilities capabilities.Capabilities, responseType model.ResponseType, name string, initPolling func(string) error, callback func(int64, int), endPolling func()) (string, error) {
	authRes, err := my.InitAuthorizationFlow(issuer, restrictions, capabilities, subtokenCapabilities, responseType, name)
	if err != nil {
		return "", err
	}
	if err = initPolling(authRes.AuthorizationURL); err != nil {
		return "", err
	}
	tok, err := my.Poll(authRes.PollingInfo, callback)
	if err == nil {
		endPolling()
	}
	return tok, err
}

func (my *Mytoken) InitAuthorizationFlow(issuer string, restrictions restrictions.Restrictions, capabilities, subtokenCapabilities capabilities.Capabilities, responseType model.ResponseType, name string) (*pkg.AuthCodeFlowResponse, error) {
	req := pkg.AuthCodeFlowRequest{
		OIDCFlowRequest: pkg.OIDCFlowRequest{
			Issuer:               issuer,
			GrantType:            model.GrantTypeOIDCFlow,
			OIDCFlow:             model.OIDCFlowAuthorizationCode,
			Restrictions:         restrictions,
			Capabilities:         capabilities,
			SubtokenCapabilities: subtokenCapabilities,
			Name:                 name,
			ResponseType:         responseType,
		},
		RedirectType: "native",
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&pkg.AuthCodeFlowResponse{}).SetError(&model.APIError{}).Post(my.SuperTokenEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError("error while sending http request", err)
	}
	if e := resp.Error(); e != nil {
		if errRes := e.(*model.APIError); errRes != nil && errRes.Error != "" {
			return nil, &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	authRes, ok := resp.Result().(*pkg.AuthCodeFlowResponse)
	if !ok {
		return nil, &MytokenError{
			err: "unexpected response from mytoken server",
		}
	}
	return authRes, nil
}

func (my *Mytoken) Poll(res pkg.PollingInfo, callback func(int64, int)) (string, error) {
	expires := time.Now().Add(time.Duration(res.PollingCodeExpiresIn) * time.Second)
	interval := res.PollingInterval
	if interval == 0 {
		interval = 5
	}
	tick := time.NewTicker(time.Duration(interval) * time.Second)
	defer tick.Stop()
	i := 0
	for t := range tick.C {
		if t.After(expires) {
			break
		}
		tok, set, err := my.PollOnce(res.PollingCode)
		if err != nil {
			return "", err
		}
		if set {
			return tok, nil
		}
		callback(res.PollingInterval, i)
		i++
	}
	return "", fmt.Errorf("polling code expired")
}

func (my *Mytoken) PollOnce(pollingCode string) (string, bool, error) {
	req := pkg.PollingCodeRequest{
		GrantType:   model.GrantTypePollingCode,
		PollingCode: pollingCode,
	}

	tok, err := my.GetSuperToken(req)
	if err == nil {
		return tok, true, nil
	}
	var myErr *MytokenError
	if errors.As(err, &myErr) {
		if myErr.err == model.ErrorAuthorizationPending {
			err = nil
		}
	}
	return tok, false, err
}
