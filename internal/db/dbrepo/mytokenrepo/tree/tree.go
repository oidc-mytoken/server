package tree

import (
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/utils/unixtime"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

// MytokenEntry holds the information of a MytokenEntry as stored in the
// database
type MytokenEntry struct {
	api.MytokenEntry `json:",inline"`
	ID               mtid.MTID         `json:"-"`
	ParentID         mtid.MTID         `db:"parent_id" json:"-"`
	Name             db.NullString     `json:"name,omitempty"`
	CreatedAt        unixtime.UnixTime `db:"created" json:"created"`
	ExpiresAt        unixtime.UnixTime `db:"expires_at" json:"expires_at,omitempty"`
	MOMID            string            `db:"mom_id" json:"mom_id"`
}

// MytokenEntryTree is a tree of MytokenEntry
type MytokenEntryTree struct {
	Token    *MytokenEntry       `json:"token"`
	Children []*MytokenEntryTree `json:"children,omitempty"`
}

// Root checks if this MytokenEntry is a root token
func (ste *MytokenEntry) Root() bool {
	return !ste.ParentID.HashValid()
}

func SingleTokenEntry(rlog log.Ext1FieldLogger, tx *sqlx.Tx, tokenID mtid.MTID) (m MytokenEntry, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&m, `CALL MTokens_GetInfo(?)`, tokenID))
		},
	)
	return

}

// AllTokens returns information about all mytokens for the user linked to the passed mytoken
func AllTokens(rlog log.Ext1FieldLogger, tx *sqlx.Tx, tokenID mtid.MTID) ([]*MytokenEntryTree, error) {
	var tokens []*MytokenEntry
	if err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&tokens, `CALL MTokens_GetAllForSameUser(?)`, tokenID))
		},
	); err != nil {
		return nil, err
	}
	return tokensToTrees(tokens), nil
}

// AllTokensByUID returns information about all mytokens for a user
func AllTokensByUID(rlog log.Ext1FieldLogger, tx *sqlx.Tx, uid uint64) (tokens []*MytokenEntry, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&tokens, `CALL MTokens_GetForUser(?)`, uid))
		},
	)
	return
}

// TokenSubTree returns information about all subtokens for the passed mytoken
func TokenSubTree(rlog log.Ext1FieldLogger, tx *sqlx.Tx, tokenID mtid.MTID) (MytokenEntryTree, error) {
	var tokens []*MytokenEntry
	if err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&tokens, `CALL MTokens_GetSubtokens(?)`, tokenID))
		},
	); err != nil {
		return MytokenEntryTree{}, err
	}
	var root *MytokenEntry
	for _, t := range tokens {
		if t.ID.Hash() == tokenID.Hash() {
			root = t
			break
		}
	}
	tree, _ := tokensToTree(root, tokens)
	return *tree, nil
}

func tokensToTrees(tokens []*MytokenEntry) (trees []*MytokenEntryTree) {
	var roots []*MytokenEntry
	for i := 0; i < len(tokens); {
		t := tokens[i]
		if t.Root() {
			removeEntry(&tokens, i)
			roots = append(roots, t)
		} else {
			i++
		}
	}
	var tmp *MytokenEntryTree
	for _, r := range roots {
		tmp, tokens = tokensToTree(r, tokens)
		trees = append(trees, tmp)
	}
	return
}

func tokensToTree(root *MytokenEntry, tokens []*MytokenEntry) (*MytokenEntryTree, []*MytokenEntry) {
	tree := &MytokenEntryTree{
		Token: root,
	}
	children := popChildren(root, &tokens)
	var cTree *MytokenEntryTree
	for _, c := range children {
		cTree, tokens = tokensToTree(c, tokens)
		tree.Children = append(tree.Children, cTree)
	}
	return tree, tokens
}

func popChildren(root *MytokenEntry, tokens *[]*MytokenEntry) (children []*MytokenEntry) {
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

func removeEntry(tokens *[]*MytokenEntry, i int) { // skipcq SCC-U1000
	copy((*tokens)[i:], (*tokens)[i+1:]) // Shift r[i+1:] left one index.
	*tokens = (*tokens)[:len(*tokens)-1] // Truncate slice.
}
