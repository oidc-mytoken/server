package context

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
)

var ctx *context.Context

func init() {
	tmp := context.Background()
	ctx = &tmp
}

func Get() context.Context {
	return *ctx
}

func SetClient(client *http.Client) {
	tmp := oidc.ClientContext(*ctx, client)
	*ctx = tmp
}
