package mytokenlib

import (
	"github.com/zachmann/mytoken/internal/server/endpoints/configuration/pkg"
	"github.com/zachmann/mytoken/pkg/model"
	"github.com/zachmann/mytoken/shared/httpClient"
	"github.com/zachmann/mytoken/shared/utils"
)

type Mytoken struct {
	pkg.MytokenConfiguration
}

func NewMytokenInstance(url string) (*Mytoken, error) {
	configEndpoint := utils.CombineURLPath(url, "/.well-known/mytoken-configuration")
	resp, err := httpClient.Do().R().SetResult(&pkg.MytokenConfiguration{}).SetError(&model.APIError{}).Get(configEndpoint)
	if err != nil {
		return nil, newMytokenErrorFromError("could not connect to mytoken instance", err)
	}
	if e := resp.Error(); e != nil {
		if errRes := e.(*model.APIError); errRes != nil && len(errRes.Error) > 0 {
			return nil, &MytokenError{
				err:          errRes.Error,
				errorDetails: errRes.ErrorDescription,
			}
		}
	}
	config, ok := resp.Result().(*pkg.MytokenConfiguration)
	if !ok {
		return nil, &MytokenError{
			err: "unexpected response from mytoken server",
		}
	}
	return &Mytoken{
		MytokenConfiguration: *config,
	}, nil
}
