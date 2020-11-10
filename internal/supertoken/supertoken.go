package supertoken

import (
	"fmt"

	"github.com/zachmann/mytoken/internal/model"

	"github.com/zachmann/mytoken/internal/db/dbModels"
	"github.com/zachmann/mytoken/internal/supertoken/capabilities"
	"github.com/zachmann/mytoken/internal/supertoken/restrictions"
)

// SuperTokenEntryTree is a tree of SuperTokenEntry
type SuperTokenEntryTree struct {
	Token    dbModels.SuperTokenEntry
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

func NewSuperTokenEntryFromSuperToken(name string, parent dbModels.SuperTokenEntry, r restrictions.Restrictions, c capabilities.Capabilities, networkData model.NetworkData) (*dbModels.SuperTokenEntry, error) {
	newRestrictions := restrictions.Tighten(parent.Token.Restrictions, r)
	newCapabilities := capabilities.Tighten(parent.Token.Capabilities, c)
	ste := dbModels.NewSuperTokenEntry(name, parent.Token.OIDCSubject, parent.Token.OIDCIssuer, newRestrictions, newCapabilities, networkData)
	ste.ParentID = parent.ID.String()
	rootID := parent.ID.String()
	if !parent.Root() {
		rootID = parent.RootID
	}
	ste.RootID = rootID
	err := ste.Store("Used grant_type super_token")
	if err != nil {
		return nil, err
	}
	return ste, nil
}
