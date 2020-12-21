package supertokenrepohelper

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

	"github.com/zachmann/mytoken/internal/db"
)

// UpdateRefreshToken updates a refresh token in the database, all occurrences of the RT are updated.
func UpdateRefreshToken(tx *sqlx.Tx, oldRT, newRT string) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`UPDATE SuperTokens SET refresh_token=? WHERE refresh_token=?`, newRT, oldRT)
		return err
	})
}

// StoreShortSuperToken stores a short super token linked to the id of a SuperToken
func StoreShortSuperToken(tx *sqlx.Tx, shortToken string, stid uuid.UUID) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO ShortSuperTokens (short_token, ST_id) VALUES(?,?)`, shortToken, stid)
		return err
	})
}

// GetRefreshToken returns the refresh token for a super token id
func GetRefreshToken(stid uuid.UUID) (string, bool, error) {
	var rt string
	err := db.DB().Get(&rt, `SELECT refresh_token FROM SuperTokens WHERE id=?`, stid)
	return parseStringResult(rt, err)
}

// GetRefreshTokenByTokenString returns the refresh token for a super token jwt string
func GetRefreshTokenByTokenString(token string) (string, bool, error) {
	var rt string
	err := db.DB().Get(&rt, `SELECT refresh_token FROM SuperTokens WHERE token=?`, token)
	return parseStringResult(rt, err)
}

func parseStringResult(res string, err error) (string, bool, error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		} else {
			return "", false, err
		}
	}
	return res, true, nil
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

// RecursiveRevokeSTByTokenString revokes the passed super token as well as all children
func RecursiveRevokeSTByTokenString(tx *sqlx.Tx, token string) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`
			DELETE FROM SuperTokens WHERE id=ANY(
			WITH Recursive childs
			AS
			(
				SELECT id, parent_id FROM SuperTokens WHERE token=?
				UNION ALL
				SELECT st.id, st.parent_id FROM SuperTokens st INNER JOIN childs c WHERE st.parent_id=c.id
			)
			SELECT id
			FROM   childs
			)`, token)
		return err
	})
}

// CheckTokenRevoked takes a short super token or a normal super token and checks if it was revoked. If the token is found in the db, the super token string will be returned.
// Therefore, this function can also be used to exchange a short super token into a normal one.
func CheckTokenRevoked(token string) (string, bool, error) {
	var count int
	if err := db.DB().Get(&count, `SELECT COUNT(1) FROM SuperTokens WHERE token=?`, token); err != nil {
		return token, true, err
	}
	if count > 0 { // token was found as SuperToken
		return token, false, nil
	}
	var superToken string
	if err := db.DB().Get(&superToken, `SELECT token FROM ShortSuperTokensV WHERE short_token=?`, token); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return token, true, err
		}
	}
	return superToken, false, nil
}

// RevokeSTByTokenString revokes the passed super token but no children
func RevokeSTByTokenString(tx *sqlx.Tx, token string) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM SuperTokens WHERE token=?`, token)
		return err
	})
}

// RevokeSTByToken revokes the passed super token and depending on the recursive parameter also its children
func RevokeSTByToken(tx *sqlx.Tx, token string, recursive bool) error {
	if recursive {
		return RecursiveRevokeSTByTokenString(tx, token)
	} else {
		return RevokeSTByTokenString(tx, token)
	}
}

// CountRTOccurrences counts how many SuperTokens use the passed refresh token
func CountRTOccurrences(tx *sqlx.Tx, rt string) (count int, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		err = tx.Get(&count, `SELECT COUNT(1) FROM SuperTokens WHERE refresh_token=?`, rt)
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
		_, err := tx.Exec(`INSERT INTO TokenUsages (ST_id, restriction, usages_AT) VALUES (?, ?, 1) ON DUPLICATE KEY UPDATE usages_AT = usages_AT + 1`, stid, jsonRestriction)
		return err
	})
}

// IncreaseTokenUsageOther increases the usage count for other usages with a SuperToken and the given restriction
func IncreaseTokenUsageOther(tx *sqlx.Tx, stid uuid.UUID, jsonRestriction []byte) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO TokenUsages (ST_id, restriction, usages_other) VALUES (?, ?, 1) ON DUPLICATE KEY UPDATE usages_other = usages_other + 1`, stid, jsonRestriction)
		return err
	})
}
