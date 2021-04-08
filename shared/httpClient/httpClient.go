package httpClient

import (
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/shared/context"
)

var client *resty.Client

func init() {
	client = resty.New()
	client.SetCookieJar(nil)
	// client.SetDisableWarn(true)
	client.SetRetryCount(2)
	client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(10))
	client.SetTimeout(20 * time.Second)
	client.GetClient()
	context.SetClient(client.GetClient())
}

// Init initializes the http client
func Init(hostURL string) {
	if hostURL != "" {
		client.SetHostURL(hostURL)
	}
	if log.IsLevelEnabled(log.DebugLevel) {
		client.SetDebug(true)
	}
}

// Do returns the client, so it can be used to do requests
func Do() *resty.Client {
	return client
}
