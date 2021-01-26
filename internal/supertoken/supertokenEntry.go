package supertoken

import (
	"fmt"

	"github.com/oidc-mytoken/server/internal/server/db/dbrepo/supertokenrepo"
)

// SuperTokenEntryTree is a tree of SuperTokenEntry
type SuperTokenEntryTree struct {
	Token    supertokenrepo.SuperTokenEntry
	Children []SuperTokenEntryTree
}

func (t *SuperTokenEntryTree) print(level int) {
	for i := 0; i < 2*level; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("%s\n", t.Token.ID)
	for _, child := range t.Children {
		child.print(level + 1)
	}
}
