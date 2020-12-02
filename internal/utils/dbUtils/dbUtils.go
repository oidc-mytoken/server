package dbUtils

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	uuid "github.com/satori/go.uuid"
	"github.com/zachmann/mytoken/internal/db"
)

func GetRefreshToken(stid uuid.UUID) (string, bool, error) {
	var rt string
	err := db.DB().Get(&rt, `SELECT refresh_token FROM SuperTokens WHERE id=?`, stid)
	return parseStringResult(rt, err)
}

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

func RecursiveRevokeSTByTokenString(token string, tx *sqlx.Tx) error {
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
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}
		return token, true, err
	}
	return superToken, true, nil
}
