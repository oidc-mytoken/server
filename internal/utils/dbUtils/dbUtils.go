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
	err := db.DB().Get(&rt, `SELECT refresh_token FROM SuperTokens WHERE id=? AND revoked=false`, stid)
	return parseStringResult(rt, err)
}

func GetRefreshTokenByTokenString(token string) (string, bool, error) {
	var rt string
	err := db.DB().Get(&rt, `SELECT refresh_token FROM SuperTokens WHERE token=? AND revoked=false`, token)
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
