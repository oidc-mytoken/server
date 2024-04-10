package pkg

import (
	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

// MTEvent is type for mytoken events
type MTEvent struct {
	Event   api.Event
	Comment string
	MTID    mtid.MTID
	api.ClientMetaData
}
