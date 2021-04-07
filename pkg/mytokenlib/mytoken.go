package mytokenlib

import (
	"errors"
	"fmt"
	"time"

	"github.com/oidc-mytoken/server/internal/endpoints/token/mytoken/pkg"
	"github.com/oidc-mytoken/server/pkg/model"
	"github.com/oidc-mytoken/server/shared/httpClient"
	"github.com/oidc-mytoken/server/shared/mytoken/capabilities"
	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
	"github.com/oidc-mytoken/server/shared/mytoken/token"
)

func (my *MytokenProvider) GetMytoken(req interface{}) (string, error) {
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&pkg.MytokenResponse{}).SetError(&model.APIError{}).Post(my.MytokenEndpoint)
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
	stRes, ok := resp.Result().(*pkg.MytokenResponse)
	if !ok {
		return "", &MytokenError{
			err: "unexpected response from mytoken server",
		}
	}
	return stRes.Mytoken, nil
}

func (my *MytokenProvider) GetMytokenByMytoken(mytoken, issuer string, restrictions restrictions.Restrictions, capabilities, subtokenCapabilities capabilities.Capabilities, responseType model.ResponseType, name string) (string, error) {
	req := pkg.MytokenFromMytokenRequest{
		Issuer:               issuer,
		GrantType:            model.GrantTypeMytoken,
		Mytoken:              token.Token(mytoken),
		Restrictions:         restrictions,
		Capabilities:         capabilities,
		SubtokenCapabilities: subtokenCapabilities,
		Name:                 name,
		ResponseType:         responseType,
	}
	return my.GetMytoken(req)
}

func (my *MytokenProvider) GetMytokenByTransferCode(transferCode string) (string, error) {
	req := pkg.ExchangeTransferCodeRequest{
		GrantType:    model.GrantTypeTransferCode,
		TransferCode: transferCode,
	}
	return my.GetMytoken(req)
}

type PollingCallbacks struct {
	Init     func(string) error
	Callback func(int64, int)
	End      func()
}

func (my *MytokenProvider) GetMytokenByAuthorizationFlow(issuer string, restrictions restrictions.Restrictions, capabilities, subtokenCapabilities capabilities.Capabilities, responseType model.ResponseType, name string, callbacks PollingCallbacks) (string, error) {
	authRes, err := my.InitAuthorizationFlow(issuer, restrictions, capabilities, subtokenCapabilities, responseType, name)
	if err != nil {
		return "", err
	}
	if err = callbacks.Init(authRes.AuthorizationURL); err != nil {
		return "", err
	}
	tok, err := my.Poll(authRes.PollingInfo, callbacks.Callback)
	if err == nil {
		callbacks.End()
	}
	return tok, err
}

func (my *MytokenProvider) InitAuthorizationFlow(issuer string, restrictions restrictions.Restrictions, capabilities, subtokenCapabilities capabilities.Capabilities, responseType model.ResponseType, name string) (*pkg.AuthCodeFlowResponse, error) {
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
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&pkg.AuthCodeFlowResponse{}).SetError(&model.APIError{}).Post(my.MytokenEndpoint)
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

func (my *MytokenProvider) Poll(res pkg.PollingInfo, callback func(int64, int)) (string, error) {
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

func (my *MytokenProvider) PollOnce(pollingCode string) (string, bool, error) {
	req := pkg.PollingCodeRequest{
		GrantType:   model.GrantTypePollingCode,
		PollingCode: pollingCode,
	}

	tok, err := my.GetMytoken(req)
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
