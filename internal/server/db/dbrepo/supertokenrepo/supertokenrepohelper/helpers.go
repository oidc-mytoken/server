package supertokenrepohelper

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

	"github.com/zachmann/mytoken/internal/server/db"
	"github.com/zachmann/mytoken/internal/server/supertoken/token"
	"github.com/zachmann/mytoken/internal/server/utils/hashUtils"
	"github.com/zachmann/mytoken/internal/utils/cryptUtils"
)

// UpdateRefreshToken updates a refresh token in the database, all occurrences of the RT are updated.
func UpdateRefreshToken(tx *sqlx.Tx, oldRT, newRT, jwt string) error {
	oldRTCrypt, err := cryptUtils.AES256Encrypt(oldRT, jwt)
	if err != nil {
		return err
	}
	newRTCrypt, err := cryptUtils.AES256Encrypt(newRT, jwt)
	if err != nil {
		return err
	}
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`UPDATE SuperTokens SET refresh_token=? WHERE refresh_token=?`, newRTCrypt, oldRTCrypt)
		return err
	})
}

// GetRefreshToken returns the refresh token for a super token id
func GetRefreshToken(stid uuid.UUID, token token.Token) (string, bool, error) {
	var rtCrypt string
	found, err := parseError(db.DB().Get(&rtCrypt, `SELECT refresh_token FROM SuperTokens WHERE id=?`, stid))
	if !found {
		return "", found, err
	}
	rt, err := cryptUtils.AES256Decrypt(rtCrypt, string(token))
	return rt, true, err
}

func parseError(err error) (bool, error) {
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
func GetSTParentID(stid uuid.UUID) (string, bool, error) {
	var parentID sql.NullString
	if err := db.DB().Get(&parentID, `SELECT parent_id FROM SuperTokens WHERE id=?`, stid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		} else {
			return "", false, err
		}
	}
	return parentID.String, true, nil
}

// GetSTRootID returns the id of the root super token of the passed super token id
func GetSTRootID(stid uuid.UUID) (string, bool, error) {
	var rootID sql.NullString
	if err := db.DB().Get(&rootID, `SELECT root_id FROM SuperTokens WHERE id=?`, stid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		} else {
			return "", false, err
		}
	}
	return rootID.String, true, nil
}

// recursiveRevokeST revokes the passed super token as well as all children
func recursiveRevokeST(tx *sqlx.Tx, id uuid.UUID) error {
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
func CheckTokenRevoked(id uuid.UUID) (bool, error) {
	var count int
	if err := db.DB().Get(&count, `SELECT COUNT(1) FROM SuperTokens WHERE id=?`, id); err != nil {
		return true, err
	}
	if count > 0 { // token was found as SuperToken
		return false, nil
	}
	return true, nil
}

// revokeST revokes the passed super token but no children
func revokeST(tx *sqlx.Tx, id uuid.UUID) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM SuperTokens WHERE id=?`, id)
		return err
	})
}

// RevokeST revokes the passed super token and depending on the recursive parameter also its children
func RevokeST(tx *sqlx.Tx, id uuid.UUID, recursive bool) error {
	if recursive {
		return recursiveRevokeST(tx, id)
	} else {
		return revokeST(tx, id)
	}
}

// CountRTOccurrences counts how many SuperTokens use the passed refresh token
func CountRTOccurrences(tx *sqlx.Tx, rt string) (count int, err error) {
	var rtHash string
	rtHash = hashUtils.SHA512Str([]byte(rt))
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		err = tx.Get(&count, `SELECT COUNT(1) FROM SuperTokens WHERE rt_hash=?`, rtHash)
		return err
	})
	return
}

// GetTokenUsagesAT returns how often a SuperToken was used with a specific restriction to obtain an access token
func GetTokenUsagesAT(tx *sqlx.Tx, stid uuid.UUID, restrictionHash string) (usages *int64, err error) {
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
func GetTokenUsagesOther(tx *sqlx.Tx, stid uuid.UUID, restrictionHash string) (usages *int64, err error) {
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
func IncreaseTokenUsageAT(tx *sqlx.Tx, stid uuid.UUID, jsonRestriction []byte) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO TokenUsages (ST_id, restriction, restriction_hash, usages_AT) VALUES (?, ?, ?, 1) ON DUPLICATE KEY UPDATE usages_AT = usages_AT + 1`, stid, jsonRestriction, hashUtils.SHA512Str(jsonRestriction))
		return err
	})
}

// IncreaseTokenUsageOther increases the usage count for other usages with a SuperToken and the given restriction
func IncreaseTokenUsageOther(tx *sqlx.Tx, stid uuid.UUID, jsonRestriction []byte) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO TokenUsages (ST_id, restriction, restriction_hash, usages_other) VALUES (?, ?, ?, 1) ON DUPLICATE KEY UPDATE usages_other = usages_other + 1`, stid, jsonRestriction, hashUtils.SHA512Str(jsonRestriction))
		return err
	})
}
