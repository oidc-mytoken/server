package mytokenrepohelper

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/encryptionkeyrepo"
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
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
		return errors.WithStack(tx.Get(&parentID, `SELECT parent_id FROM MTokens WHERE id=?`, myID))
	}))
	return parentID.String, found, err
}

// GetMTRootID returns the id of the root mytoken of the passed mytoken id
func GetMTRootID(id mtid.MTID) (mtid.MTID, bool, error) {
	var rootID mtid.MTID
	found, err := ParseError(db.Transact(func(tx *sqlx.Tx) error {
		return errors.WithStack(tx.Get(&rootID, `SELECT root_id FROM MTokens WHERE id=?`, id))
	}))
	return rootID, found, err
}

// recursiveRevokeMT revokes the passed mytoken as well as all children
func recursiveRevokeMT(tx *sqlx.Tx, id mtid.MTID) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		var effectedMTIDs []mtid.MTID
		if err := errors.WithStack(tx.Select(&effectedMTIDs,
			`WITH Recursive childs AS (
                  SELECT id, parent_id FROM MTokens WHERE id=?
                  UNION ALL
                  SELECT mt.id, mt.parent_id FROM MTokens mt INNER JOIN childs c WHERE mt.parent_id=c.id
                ) SELECT id FROM   childs`,
			id)); err != nil {
			return err
		}
		query, args, err := sqlx.In(
			`DELETE FROM EncryptionKeys WHERE id=ANY(SELECT key_id FROM RT_EncryptionKeys WHERE MT_id IN (?))`,
			effectedMTIDs)
		if err != nil {
			return errors.WithStack(err)
		}
		if _, err = tx.Exec(query, args...); err != nil {
			return errors.WithStack(err)
		}
		query, args, err = sqlx.In(`DELETE FROM MTokens WHERE id IN (?)`, effectedMTIDs)
		if err != nil {
			return errors.WithStack(err)
		}
		_, err = tx.Exec(query, args...)
		return errors.WithStack(err)
	})
}

// CheckTokenRevoked checks if a Mytoken has been revoked. If it is a rotating mytoken and auto_revoke is enabled for
// this token, it might get triggered.
func CheckTokenRevoked(tx *sqlx.Tx, id mtid.MTID, seqno uint64, rot *api.Rotation) (bool, error) {
	if rot == nil {
		return checkTokenRevoked(tx, id, seqno)
	}
	var revoked bool
	var err error
	if rot.Lifetime > 0 {
		revoked, err = checkRotatingTokenRevoked(tx, id, seqno, rot.Lifetime)
	} else {
		revoked, err = checkTokenRevoked(tx, id, seqno)
	}
	if err != nil {
		return revoked, err
	}
	if !revoked || !rot.AutoRevoke {
		return revoked, nil
	}
	// At this point we know, that the token is not valid, we now check if it is not valid because of the seqno.
	idFound, err := checkTokenID(tx, id)
	if err != nil {
		return true, err
	}
	if !idFound {
		return true, nil
	}
	err = RevokeMT(tx, id, true)
	return true, err
}

func checkTokenRevoked(tx *sqlx.Tx, id mtid.MTID, seqno uint64) (bool, error) {
	var count int
	if err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return errors.WithStack(tx.Get(&count, `SELECT COUNT(1) FROM MTokens WHERE id=? AND seqno=?`, id, seqno))
	}); err != nil {
		return true, err
	}
	if count > 0 { // token was found as Mytoken
		return false, nil
	}
	return true, nil
}

func checkTokenID(tx *sqlx.Tx, id mtid.MTID) (bool, error) {
	var count int
	if err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return errors.WithStack(tx.Get(&count, `SELECT COUNT(1) FROM MTokens WHERE id=?`, id))
	}); err != nil {
		return false, err
	}
	return count > 0, nil
}

func checkRotatingTokenRevoked(tx *sqlx.Tx, id mtid.MTID, seqno, rotationLifetime uint64) (bool, error) {
	var count int
	if err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return errors.WithStack(tx.Get(&count,
			`SELECT COUNT(1) FROM MTokens WHERE id=? AND seqno=?
                               AND TIMESTAMPADD(SECOND, ?, last_rotated) >= CURRENT_TIMESTAMP()`,
			id, seqno, rotationLifetime))
	}); err != nil {
		return true, err
	}
	if count > 0 { // token was found as Mytoken
		return false, nil
	}
	return true, nil
}

// UpdateSeqNo updates the sequence number of a mytoken, i.e. it rotates the mytoken. Don't forget to update the
// encryption key
func UpdateSeqNo(tx *sqlx.Tx, id mtid.MTID, seqno uint64) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`UPDATE MTokens SET seqno=?, last_rotated=current_timestamp() WHERE id=?`, seqno, id)
		return errors.WithStack(err)
	})
}

// revokeMT revokes the passed mytoken but no children
func revokeMT(tx *sqlx.Tx, id mtid.MTID) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		if err := encryptionkeyrepo.DeleteEncryptionKey(tx, id); err != nil {
			return err
		}
		_, err := tx.Exec(`DELETE FROM MTokens WHERE id=?`, id)
		return errors.WithStack(err)
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
		return errors.WithStack(tx.Get(&usageCount,
			`SELECT usages_AT FROM TokenUsages WHERE restriction_hash=? AND MT_id=?`,
			restrictionHash, myID))
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

// GetTokenUsagesOther returns how often a Mytoken was used with a specific restriction to do something else than
// obtaining an access token
func GetTokenUsagesOther(tx *sqlx.Tx, myID mtid.MTID, restrictionHash string) (usages *int64, err error) {
	var usageCount int64
	if err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return errors.WithStack(tx.Get(&usageCount,
			`SELECT usages_other FROM TokenUsages WHERE restriction_hash=? AND MT_id=?`,
			restrictionHash, myID))
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
		_, err := tx.Exec(
			`INSERT INTO TokenUsages (MT_id, restriction, restriction_hash, usages_AT) VALUES (?, ?, ?, 1)
                      ON DUPLICATE KEY UPDATE usages_AT = usages_AT + 1`,
			myID, jsonRestriction, hashUtils.SHA512Str(jsonRestriction))
		return errors.WithStack(err)
	})
}

// IncreaseTokenUsageOther increases the usage count for other usages with a Mytoken and the given restriction
func IncreaseTokenUsageOther(tx *sqlx.Tx, myID mtid.MTID, jsonRestriction []byte) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO TokenUsages (MT_id, restriction, restriction_hash, usages_other) VALUES (?, ?, ?, 1) 
                      ON DUPLICATE KEY UPDATE usages_other = usages_other + 1`,
			myID, jsonRestriction, hashUtils.SHA512Str(jsonRestriction))
		return errors.WithStack(err)
	})
}
