package mytokenlib

import (
	"github.com/oidc-mytoken/server/pkg/api/v0"
	"github.com/oidc-mytoken/server/shared/httpClient"
)

func (my *MytokenProvider) TokeninfoIntrospect(mytoken string) (*api.TokeninfoIntrospectResponse, error) {
	req := api.TokenInfoRequest{
		Action:  api.TokeninfoActionIntrospect,
		Mytoken: mytoken,
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&api.TokeninfoIntrospectResponse{}).SetError(&api.APIError{}).Post(my.TokeninfoEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError(errorWhileHttp, err)
	}
	if eRes := resp.Error(); eRes != nil {
		if errRes := eRes.(*api.APIError); errRes != nil && errRes.Error != "" {
			return nil, &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	res, ok := resp.Result().(*api.TokeninfoIntrospectResponse)
	if !ok {
		return nil, &MytokenError{
			err: unexpectedResponse,
		}
	}
	return res, nil
}
func (my *MytokenProvider) TokeninfoHistory(mytoken string) (*api.TokeninfoHistoryResponse, error) {
	req := api.TokenInfoRequest{
		Action:  api.TokeninfoActionEventHistory,
		Mytoken: mytoken,
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&api.TokeninfoHistoryResponse{}).SetError(&api.APIError{}).Post(my.TokeninfoEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError(errorWhileHttp, err)
	}
	if eRes := resp.Error(); eRes != nil {
		if errRes := eRes.(*api.APIError); errRes != nil && errRes.Error != "" {
			return nil, &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	res, ok := resp.Result().(*api.TokeninfoHistoryResponse)
	if !ok {
		return nil, &MytokenError{
			err: unexpectedResponse,
		}
	}
	return res, nil
}
func (my *MytokenProvider) TokeninfoSubtokens(mytoken string) (*api.TokeninfoTreeResponse, error) {
	req := api.TokenInfoRequest{
		Action:  api.TokeninfoActionSubtokenTree,
		Mytoken: mytoken,
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&api.TokeninfoTreeResponse{}).SetError(&api.APIError{}).Post(my.TokeninfoEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError(errorWhileHttp, err)
	}
	if eRes := resp.Error(); eRes != nil {
		if errRes := eRes.(*api.APIError); errRes != nil && errRes.Error != "" {
			return nil, &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	res, ok := resp.Result().(*api.TokeninfoTreeResponse)
	if !ok {
		return nil, &MytokenError{
			err: unexpectedResponse,
		}
	}
	return res, nil
}
func (my *MytokenProvider) TokeninfoListMytokens(mytoken string) (*api.TokeninfoListResponse, error) {
	req := api.TokenInfoRequest{
		Action:  api.TokeninfoActionListMytokens,
		Mytoken: mytoken,
	}
	resp, err := httpClient.Do().R().SetBody(req).SetResult(&api.TokeninfoListResponse{}).SetError(&api.APIError{}).Post(my.TokeninfoEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError(errorWhileHttp, err)
	}
	if eRes := resp.Error(); eRes != nil {
		if errRes := eRes.(*api.APIError); errRes != nil && errRes.Error != "" {
			return nil, &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	res, ok := resp.Result().(*api.TokeninfoListResponse)
	if !ok {
		return nil, &MytokenError{
			err: unexpectedResponse,
		}
	}
	return res, nil
}
