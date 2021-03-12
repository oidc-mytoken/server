package refreshtokenrepo

import (
	"encoding/base64"

	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	helper "github.com/oidc-mytoken/server/internal/db/dbrepo/supertokenrepo/supertokenrepohelper"
	"github.com/oidc-mytoken/server/shared/supertoken/pkg/stid"
	"github.com/oidc-mytoken/server/shared/utils/cryptUtils"
)

// UpdateRefreshToken updates a refresh token in the database, all occurrences of the RT are updated.
func UpdateRefreshToken(tx *sqlx.Tx, tokenID stid.STID, newRT, jwt string) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		key, rtID, err := GetEncryptionKey(tx, tokenID, jwt)
		if err != nil {
			return err
		}
		updatedRT, err := cryptUtils.AESEncrypt(newRT, key)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`UPDATE RefreshTokens SET rt=? WHERE id=?`, updatedRT, rtID)
		return err
	})
}

func GetEncryptionKey(tx *sqlx.Tx, tokenID stid.STID, jwt string) ([]byte, uint64, error) {
	var key []byte
	var rtID uint64
	err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		var res struct {
			EncryptedKey encryptionKey `db:"encryption_key"`
			ID           uint64        `db:"rt_id"`
		}
		if err := tx.Get(&res, `SELECT encryption_key, rt_id FROM MyTokens WHERE id=?`, tokenID); err != nil {
			return err
		}
		rtID = res.ID
		tmp, err := res.EncryptedKey.decrypt(jwt)
		key = tmp
		return err
	})
	return key, rtID, err
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
func GetRefreshToken(tx *sqlx.Tx, stid stid.STID, jwt string) (string, bool, error) {
	var rt rtStruct
	found, err := helper.ParseError(db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&rt, `SELECT refresh_token, encryption_key FROM MyTokens WHERE id=?`, stid)
	}))
	if !found {
		return "", found, err
	}
	plainRT, err := rt.decrypt(jwt)
	return plainRT, true, err
}

// DeleteRefreshToken deletes a refresh token
func DeleteRefreshToken(tx *sqlx.Tx, rtID uint64) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM RefreshTokens WHERE id=?`, rtID)
		return err
	})
}

// CountRTOccurrences counts how many SuperTokens use the passed refresh token
func CountRTOccurrences(tx *sqlx.Tx, rtID uint64) (count int, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&count, `SELECT COUNT(1) FROM SuperTokens WHERE rt_id=?`, rtID)
	})
	return
}

// GetRTID returns the refresh token id for a supertoken
func GetRTID(tx *sqlx.Tx, mytID stid.STID) (rtID uint64, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&rtID, `SELECT rt_ID FROM SuperTokens WHERE id=?`, mytID)
	})
	return
}
