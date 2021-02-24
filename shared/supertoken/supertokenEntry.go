package supertoken

import (
	"github.com/oidc-mytoken/server/internal/db/dbrepo/supertokenrepo"
)

// SuperTokenEntryTree is a tree of SuperTokenEntry
type SuperTokenEntryTree struct {
	Token    supertokenrepo.SuperTokenEntry
	Children []SuperTokenEntryTree
}
