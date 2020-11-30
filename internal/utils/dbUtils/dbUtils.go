package dbUtils

import (
	"database/sql"
	"errors"

	uuid "github.com/satori/go.uuid"
	"github.com/zachmann/mytoken/internal/db"
)

func GetRefreshToken(stid uuid.UUID) (string, bool, error) {
	var rt string
	if err := db.DB().Get(&rt, `SELECT refresh_token FROM SuperTokens WHERE id=? AND revoked=false`, stid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		} else {
			return "", false, err
		}
	}
	return rt, true, nil
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
