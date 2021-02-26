package tree

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/shared/supertoken/pkg/stid"
	"github.com/oidc-mytoken/server/shared/utils/unixtime"
)

// SuperTokenEntry holds the information of a SuperTokenEntry as stored in the
// database
type SuperTokenEntry struct {
	ID        stid.STID         `json:"-"`
	ParentID  stid.STID         `db:"parent_id" json:"-"`
	RootID    stid.STID         `db:"root_id" json:"-"`
	Name      db.NullString     `json:"name,omitempty"`
	CreatedAt unixtime.UnixTime `db:"created" json:"created"`
	model.ClientMetaData
}

// Root checks if this SuperTokenEntry is a root token
func (ste *SuperTokenEntry) Root() bool {
	if ste.ID.Hash() == ste.RootID.Hash() {
		return true
	}
	return !ste.RootID.HashValid()
}

// SuperTokenEntryTree is a tree of SuperTokenEntry
type SuperTokenEntryTree struct {
	Token    SuperTokenEntry       `json:"token"`
	Children []SuperTokenEntryTree `json:"children,omitempty"`
}

func (stet SuperTokenEntryTree) print(depth int) {
	fmt.Printf("%s%s - %s - %s\n", strings.Repeat(" ", depth*4), stet.Token.Name.String, stet.Token.CreatedAt.Time().String(), stet.Token.IP)
	for _, c := range stet.Children {
		c.print(depth + 1)
	}
}

func GetUserID(tx *sqlx.Tx, tokenID stid.STID) (uid int64, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&uid, `SELECT user_id FROM SuperTokens WHERE id=? ORDER BY name`, tokenID)
	})
	return
}

func AllTokens(tx *sqlx.Tx, tokenID stid.STID) (trees []SuperTokenEntryTree, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		uid, e := GetUserID(tx, tokenID)
		if e != nil {
			return e
		}
		trees, err = AllTokensForUser(tx, uid)
		return err
	})
	return
}
func AllTokensForUser(tx *sqlx.Tx, uid int64) ([]SuperTokenEntryTree, error) {
	var tokens []SuperTokenEntry
	if err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Select(&tokens, `SELECT id, parent_id, root_id, name, created, ip_created AS ip FROM SuperTokens WHERE user_id=?`, uid)
	}); err != nil {
		return nil, err
	}
	return tokensToTrees(tokens), nil
}
func subtokens(tx *sqlx.Tx, rootID stid.STID) ([]SuperTokenEntry, error) {
	var tokens []SuperTokenEntry
	err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Select(&tokens, `SELECT id, parent_id, root_id, name, created, ip_created AS ip FROM SuperTokens WHERE root_id=?`, rootID)
	})
	return tokens, err
}
func TokenSubTree(tx *sqlx.Tx, tokenID stid.STID) (SuperTokenEntryTree, error) {
	var tokens []SuperTokenEntry
	var root SuperTokenEntry
	if err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		var err error
		if err = tx.Get(&root, `SELECT id, parent_id, root_id, name, created, ip_created AS ip FROM SuperTokens WHERE id=?`, tokenID); err != nil {
			return err
		}
		if root.Root() {
			root.RootID = root.ID
		}
		tokens, err = subtokens(tx, root.RootID)
		return err
	}); err != nil {
		return SuperTokenEntryTree{}, err
	}
	tree, _ := tokensToTree(root, tokens)
	return tree, nil
}

func tokensToTrees(tokens []SuperTokenEntry) (trees []SuperTokenEntryTree) {
	var roots []SuperTokenEntry
	for i := 0; i < len(tokens); {
		t := tokens[i]
		if t.Root() {
			removeEntry(&tokens, i)
			roots = append(roots, t)
		} else {
			i++
		}
	}
	var tmp SuperTokenEntryTree
	for _, r := range roots {
		tmp, tokens = tokensToTree(r, tokens)
		trees = append(trees, tmp)
	}
	return
}

func tokensToTree(root SuperTokenEntry, tokens []SuperTokenEntry) (SuperTokenEntryTree, []SuperTokenEntry) {
	tree := SuperTokenEntryTree{
		Token: root,
	}
	children := popChildren(root, &tokens)
	var cTree SuperTokenEntryTree
	for _, c := range children {
		cTree, tokens = tokensToTree(c, tokens)
		tree.Children = append(tree.Children, cTree)
	}
	return tree, tokens
}

func popChildren(root SuperTokenEntry, tokens *[]SuperTokenEntry) (children []SuperTokenEntry) {
	i := 0
	for i < len(*tokens) {
		t := (*tokens)[i]
		if t.ParentID == root.ID {
			removeEntry(tokens, i)
			children = append(children, t)
		} else {
			i++
		}
	}
	return
}

func removeEntry(tokens *[]SuperTokenEntry, i int) { // skipcq SCC-U1000
	copy((*tokens)[i:], (*tokens)[i+1:]) // Shift r[i+1:] left one index.
	*tokens = (*tokens)[:len(*tokens)-1] // Truncate slice.
}
