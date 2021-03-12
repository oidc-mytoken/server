package supertokenrepohelper

import (
	"database/sql"
	"encoding/base64"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
	"github.com/oidc-mytoken/server/shared/supertoken/pkg/stid"
	"github.com/oidc-mytoken/server/shared/utils/cryptUtils"
)

// UpdateRefreshToken updates a refresh token in the database, all occurrences of the RT are updated.
func UpdateRefreshToken(tx *sqlx.Tx, tokenID stid.STID, newRT, jwt string) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		key, rtHash, err := GetEncryptionKey(tx, tokenID, jwt)
		if err != nil {
			return err
		}
		updatedRT, err := cryptUtils.AESEncrypt(newRT, key)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`UPDATE RefreshTokens SET rt=? WHERE hash=?`, updatedRT, rtHash)
		return err
	})
}

func GetEncryptionKey(tx *sqlx.Tx, tokenID stid.STID, jwt string) ([]byte, string, error) {
	var key []byte
	var hash string
	err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		var res struct {
			EncryptedKey encryptionKey `db:"encryption_key"`
			Hash         string        `db:"rt_hash"`
		}
		if err := tx.Get(&res, `SELECT encryption_key, rt_hash FROM MyTokens WHERE id=?`, tokenID); err != nil {
			return err
		}
		hash = res.Hash
		tmp, err := res.EncryptedKey.decrypt(jwt)
		key = tmp
		return err
	})
	return key, hash, err
}

type encryptionKey string

func (k encryptionKey) decrypt(jwt string) ([]byte, error) {
	decryptedKey, err := cryptUtils.AES256Decrypt(string(k), jwt)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(decryptedKey)
}

type rtStruct struct {
	RT  string        `db:"refresh_token"`
	Key encryptionKey `db:"encryption_key"`
}

func (rt rtStruct) decrypt(jwt string) (string, error) {
	key, err := rt.Key.decrypt(jwt)
	if err != nil {
		return "", err
	}
	return cryptUtils.AESDecrypt(rt.RT, key)
}

// GetRefreshToken returns the refresh token for a super token id
func GetRefreshToken(stid stid.STID, jwt string) (string, bool, error) {
	var rt rtStruct
	found, err := parseError(db.Transact(func(tx *sqlx.Tx) error {
		return tx.Get(&rt, `SELECT refresh_token, encryption_key FROM MyTokens WHERE id=?`, stid)
	}))
	if !found {
		return "", found, err
	}
	plainRT, err := rt.decrypt(jwt)
	return plainRT, true, err
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
func GetSTParentID(stid stid.STID) (string, bool, error) {
	var parentID sql.NullString
	found, err := parseError(db.Transact(func(tx *sqlx.Tx) error {
		return tx.Get(&parentID, `SELECT parent_id FROM SuperTokens WHERE id=?`, stid)
	}))
	return parentID.String, found, err
}

// GetSTRootID returns the id of the root super token of the passed super token id
func GetSTRootID(id stid.STID) (stid.STID, bool, error) {
	var rootID stid.STID
	found, err := parseError(db.Transact(func(tx *sqlx.Tx) error {
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

// CountRTOccurrences counts how many SuperTokens use the passed refresh token
func CountRTOccurrences(tx *sqlx.Tx, rt string) (count int, err error) {
	var rtHash string = hashUtils.SHA512Str([]byte(rt))
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		err = tx.Get(&count, `SELECT COUNT(1) FROM SuperTokens WHERE rt_hash=?`, rtHash)
		return err
	})
	return
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
