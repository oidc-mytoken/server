package mytokenrepohelper

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/mytoken/rotation"
)

// ParseError parses the passed error for a sql.ErrNoRows
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

// GetMTParentID returns the id of the parent mytoken of the passed mytoken id
func GetMTParentID(myID mtid.MTID) (string, bool, error) {
	var parentID sql.NullString
	found, err := ParseError(db.Transact(func(tx *sqlx.Tx) error {
		return tx.Get(&parentID, `SELECT parent_id FROM MTokens WHERE id=?`, myID)
	}))
	return parentID.String, found, err
}

// GetMTRootID returns the id of the root mytoken of the passed mytoken id
func GetMTRootID(id mtid.MTID) (mtid.MTID, bool, error) {
	var rootID mtid.MTID
	found, err := ParseError(db.Transact(func(tx *sqlx.Tx) error {
		return tx.Get(&rootID, `SELECT root_id FROM MTokens WHERE id=?`, id)
	}))
	return rootID, found, err
}

// recursiveRevokeMT revokes the passed mytoken as well as all children
func recursiveRevokeMT(tx *sqlx.Tx, id mtid.MTID) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`
			DELETE FROM MTokens WHERE id=ANY(
			WITH Recursive childs
			AS
			(
				SELECT id, parent_id FROM MTokens WHERE id=?
				UNION ALL
				SELECT mt.id, mt.parent_id FROM MTokens mt INNER JOIN childs c WHERE mt.parent_id=c.id
			)
			SELECT id
			FROM   childs
			)`, id)
		return err
	})
}

// CheckTokenRevoked checks if a Mytoken has been revoked.
func CheckTokenRevoked(tx *sqlx.Tx, id mtid.MTID, seqno uint64, rot *rotation.Rotation) (bool, error) {
	if rot != nil && rot.Lifetime > 0 {
		return checkRotatingTokenRevoked(tx, id, seqno, rot.Lifetime)
	}
	return checkTokenRevoked(tx, id, seqno)
}

func checkTokenRevoked(tx *sqlx.Tx, id mtid.MTID, seqno uint64) (bool, error) {
	var count int
	if err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&count, `SELECT COUNT(1) FROM MTokens WHERE id=? AND seqno=?`, id, seqno)
	}); err != nil {
		return true, err
	}
	if count > 0 { // token was found as Mytoken
		return false, nil
	}
	return true, nil
}

func checkRotatingTokenRevoked(tx *sqlx.Tx, id mtid.MTID, seqno, rotationLifetime uint64) (bool, error) {
	var count int
	if err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&count, `SELECT COUNT(1) FROM MTokens WHERE id=? AND seqno=? AND TIMESTAMPADD(SECOND, ?, last_rotated) >= CURRENT_TIMESTAMP()`, id, seqno, rotationLifetime)
	}); err != nil {
		return true, err
	}
	if count > 0 { // token was found as Mytoken
		return false, nil
	}
	return true, nil
}

// revokeMT revokes the passed mytoken but no children
func revokeMT(tx *sqlx.Tx, id mtid.MTID) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM MTokens WHERE id=?`, id)
		return err
	})
}

// RevokeMT revokes the passed mytoken and depending on the recursive parameter also its children
func RevokeMT(tx *sqlx.Tx, id mtid.MTID, recursive bool) error {
	if recursive {
		return recursiveRevokeMT(tx, id)
	} else {
		return revokeMT(tx, id)
	}
}

// GetTokenUsagesAT returns how often a Mytoken was used with a specific restriction to obtain an access token
func GetTokenUsagesAT(tx *sqlx.Tx, myID mtid.MTID, restrictionHash string) (usages *int64, err error) {
	var usageCount int64
	if err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&usageCount, `SELECT usages_AT FROM TokenUsages WHERE restriction_hash=? AND MT_id=?`, restrictionHash, myID)
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

// GetTokenUsagesOther returns how often a Mytoken was used with a specific restriction to do something else than obtaining an access token
func GetTokenUsagesOther(tx *sqlx.Tx, myID mtid.MTID, restrictionHash string) (usages *int64, err error) {
	var usageCount int64
	if err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&usageCount, `SELECT usages_other FROM TokenUsages WHERE restriction_hash=? AND MT_id=?`, restrictionHash, myID)
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

// IncreaseTokenUsageAT increases the usage count for obtaining ATs with a Mytoken and the given restriction
func IncreaseTokenUsageAT(tx *sqlx.Tx, myID mtid.MTID, jsonRestriction []byte) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO TokenUsages (MT_id, restriction, restriction_hash, usages_AT) VALUES (?, ?, ?, 1) ON DUPLICATE KEY UPDATE usages_AT = usages_AT + 1`, myID, jsonRestriction, hashUtils.SHA512Str(jsonRestriction))
		return err
	})
}

// IncreaseTokenUsageOther increases the usage count for other usages with a Mytoken and the given restriction
func IncreaseTokenUsageOther(tx *sqlx.Tx, myID mtid.MTID, jsonRestriction []byte) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO TokenUsages (MT_id, restriction, restriction_hash, usages_other) VALUES (?, ?, ?, 1) ON DUPLICATE KEY UPDATE usages_other = usages_other + 1`, myID, jsonRestriction, hashUtils.SHA512Str(jsonRestriction))
		return err
	})
}
