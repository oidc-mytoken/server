package tree

import (
	"fmt"
	"testing"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/shared/supertoken/pkg/stid"
)

func TestTree(t *testing.T) {
	config.LoadForSetup()
	db.Connect()
	tree, err := AllTokensForUser(nil, 4)
	if err != nil {
		t.Error(err)
	}
	for _, tt := range tree {
		tt.print(0)
	}
}

func str(tokens []SuperTokenEntry) string {
	s := ""
	for _, t := range tokens {
		s += " '" + t.Name.String + "'"
	}
	return s
}

func TestPopChildren(t *testing.T) {
	root := SuperTokenEntry{
		ID:   stid.New(),
		Name: db.NewNullString("root"),
	}
	tokens := []SuperTokenEntry{
		{ID: stid.New(), ParentID: root.ID, RootID: root.ID, Name: db.NewNullString("child1")},
		{ID: stid.New(), ParentID: root.ID, RootID: root.ID, Name: db.NewNullString("child2")},
		{ID: stid.New(), Name: db.NewNullString("other")},
	}
	tokens = append(tokens, SuperTokenEntry{ID: stid.New(), ParentID: tokens[0].ID, RootID: root.ID, Name: db.NewNullString("subchild")})
	expectedRest := append([]SuperTokenEntry{}, tokens[2:]...)
	expectedChildren := append([]SuperTokenEntry{}, tokens[:2]...)
	children := popChildren(root, &tokens)
	if len(tokens) != len(expectedRest) {
		t.Errorf("Rest of tokens not as expected:\nExpected: %s\nGot: %s'", str(expectedRest), str(tokens))
	} else {
		for i, tt := range tokens {
			if tt.ID != expectedRest[i].ID {
				t.Errorf("Rest of tokens not as expected:\nExpected: %s\nGot: %s'", str(expectedRest), str(tokens))
			}
		}
	}
	if len(children) != len(expectedChildren) {
		t.Errorf("Children not as expected:\nExpected: %s\nGot: %s", str(expectedChildren), str(children))
	} else {
		for i, tt := range children {
			if tt.ID != expectedChildren[i].ID {
				t.Errorf("Children not as expected:\nExpected: %s\nGot: %s", str(expectedChildren), str(children))
			}
		}
	}
}

func TestTokensToTree(t *testing.T) {
	root := SuperTokenEntry{
		ID:   stid.New(),
		Name: db.NewNullString("root"),
	}
	tokens := []SuperTokenEntry{
		{ID: stid.New(), ParentID: root.ID, RootID: root.ID, Name: db.NewNullString("child1")},
		{ID: stid.New(), ParentID: root.ID, RootID: root.ID, Name: db.NewNullString("child2")},
		{ID: stid.New(), Name: db.NewNullString("other")},
	}
	tokens = append(tokens, SuperTokenEntry{ID: stid.New(), ParentID: tokens[0].ID, RootID: root.ID, Name: db.NewNullString("subchild")})
	expectedRest := append([]SuperTokenEntry{}, tokens[2])
	expectedTree := SuperTokenEntryTree{
		Token: root,
		Children: []SuperTokenEntryTree{
			{
				Token: tokens[0],
				Children: []SuperTokenEntryTree{
					{
						Token: tokens[3],
					},
				},
			},
			{
				Token: tokens[1],
			},
		},
	}
	tree, rest := tokensToTree(root, tokens)
	if len(rest) != len(expectedRest) {
		t.Errorf("Rest of tokens not as expected:\nExpected: %s\nGot: %s'", str(expectedRest), str(rest))
	} else {
		for i, tt := range rest {
			if tt.ID != expectedRest[i].ID {
				t.Errorf("Rest of tokens not as expected:\nExpected: %s\nGot: %s'", str(expectedRest), str(rest))
			}
		}
	}
	if !compareTrees(tree, expectedTree) {
		fmt.Println("Expected Tree:")
		expectedTree.print(0)
		fmt.Println("Got Tree:")
		tree.print(0)
		t.Errorf("Treese not as expected")
	}
}

func compareTrees(a, b SuperTokenEntryTree) bool {
	if a.Token.ID != b.Token.ID {
		return false
	}
	if len(a.Children) != len(b.Children) {
		return false
	}
	for i, cc := range a.Children {
		if !compareTrees(cc, b.Children[i]) {
			return false
		}
	}
	return true
}
