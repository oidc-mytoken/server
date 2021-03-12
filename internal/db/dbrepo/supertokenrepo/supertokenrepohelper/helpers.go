package supertokenrepohelper

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
	"github.com/oidc-mytoken/server/shared/supertoken/pkg/stid"
)

func ParseError(err error) (bool, error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

// GetSTParentID returns the id of the parent super token of the passed super token id
func GetSTParentID(stid stid.STID) (string, bool, error) {
	var parentID sql.NullString
	found, err := ParseError(db.Transact(func(tx *sqlx.Tx) error {
		return tx.Get(&parentID, `SELECT parent_id FROM SuperTokens WHERE id=?`, stid)
	}))
	return parentID.String, found, err
}

// GetSTRootID returns the id of the root super token of the passed super token id
func GetSTRootID(id stid.STID) (stid.STID, bool, error) {
	var rootID stid.STID
	found, err := ParseError(db.Transact(func(tx *sqlx.Tx) error {
		return tx.Get(&rootID, `SELECT root_id FROM SuperTokens WHERE id=?`, id)
	}))
	return rootID, found, err
}

// recursiveRevokeST revokes the passed super token as well as all children
func recursiveRevokeST(tx *sqlx.Tx, id stid.STID) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`
			DELETE FROM SuperTokens WHERE id=ANY(
			WITH Recursive childs
			AS
			(
				SELECT id, parent_id FROM SuperTokens WHERE id=?
				UNION ALL
				SELECT st.id, st.parent_id FROM SuperTokens st INNER JOIN childs c WHERE st.parent_id=c.id
			)
			SELECT id
			FROM   childs
			)`, id)
		return err
	})
}

// CheckTokenRevoked checks if a SuperToken has been revoked.
func CheckTokenRevoked(id stid.STID) (bool, error) {
	var count int
	if err := db.Transact(func(tx *sqlx.Tx) error {
		return tx.Get(&count, `SELECT COUNT(1) FROM SuperTokens WHERE id=?`, id)
	}); err != nil {
		return true, err
	}
	if count > 0 { // token was found as SuperToken
		return false, nil
	}
	return true, nil
}

// revokeST revokes the passed super token but no children
func revokeST(tx *sqlx.Tx, id stid.STID) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM SuperTokens WHERE id=?`, id)
		return err
	})
}

// RevokeST revokes the passed super token and depending on the recursive parameter also its children
func RevokeST(tx *sqlx.Tx, id stid.STID, recursive bool) error {
	if recursive {
		return recursiveRevokeST(tx, id)
	} else {
		return revokeST(tx, id)
	}
}

// GetTokenUsagesAT returns how often a SuperToken was used with a specific restriction to obtain an access token
func GetTokenUsagesAT(tx *sqlx.Tx, stid stid.STID, restrictionHash string) (usages *int64, err error) {
	var usageCount int64
	if err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&usageCount, `SELECT usages_AT FROM TokenUsages WHERE restriction_hash=? AND ST_id=?`, restrictionHash, stid)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No usage entry -> was not used before -> usages=nil
			err = nil // This is fine
			return
		}
		return
	}
	usages = &usageCount
	return
}

// GetTokenUsagesOther returns how often a SuperToken was used with a specific restriction to do something else than obtaining an access token
func GetTokenUsagesOther(tx *sqlx.Tx, stid stid.STID, restrictionHash string) (usages *int64, err error) {
	var usageCount int64
	if err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&usageCount, `SELECT usages_other FROM TokenUsages WHERE restriction_hash=? AND ST_id=?`, restrictionHash, stid)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No usage entry -> was not used before -> usages=nil
			err = nil // This is fine
			return
		}
		return
	}
	usages = &usageCount
	return
}

// IncreaseTokenUsageAT increases the usage count for obtaining ATs with a SuperToken and the given restriction
func IncreaseTokenUsageAT(tx *sqlx.Tx, stid stid.STID, jsonRestriction []byte) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO TokenUsages (ST_id, restriction, restriction_hash, usages_AT) VALUES (?, ?, ?, 1) ON DUPLICATE KEY UPDATE usages_AT = usages_AT + 1`, stid, jsonRestriction, hashUtils.SHA512Str(jsonRestriction))
		return err
	})
}

// IncreaseTokenUsageOther increases the usage count for other usages with a SuperToken and the given restriction
func IncreaseTokenUsageOther(tx *sqlx.Tx, stid stid.STID, jsonRestriction []byte) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO TokenUsages (ST_id, restriction, restriction_hash, usages_other) VALUES (?, ?, ?, 1) ON DUPLICATE KEY UPDATE usages_other = usages_other + 1`, stid, jsonRestriction, hashUtils.SHA512Str(jsonRestriction))
		return err
	})
}
