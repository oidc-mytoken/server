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

// Get returns the shared context
func Get() context.Context {
	return *ctx
}

// SetClient sets the passed client into the context
func SetClient(client *http.Client) {
	tmp := oidc.ClientContext(*ctx, client)
	*ctx = tmp
}
