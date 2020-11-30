package httpClient

import (
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/zachmann/mytoken/internal/config"
)

var client *resty.Client

func init() {
	client = resty.New()
	client.SetCookieJar(nil)
	//client.SetDisableWarn(true)
	client.SetRetryCount(2)
	client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(10))
	client.SetTimeout(20 * time.Second)

}

func Init() {
	client.SetHostURL(config.Get().IssuerURL)
	if log.IsLevelEnabled(log.DebugLevel) {
		client.SetDebug(true)
	}
}

func Do() *resty.Client {
	return client
}
