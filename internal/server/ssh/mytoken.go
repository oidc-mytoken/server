package ssh

import (
	"context"

	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
)

func handleSSHMytoken(req []byte, ctx context.Context) error {
	return errors.New(api.ErrorNYI.CombinedMessage())
}
