package refreshtokenrepo

import (
	"encoding/base64"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/server/internal/db"
	helper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/utils/cryptUtils"
)

// UpdateRefreshToken updates a refresh token in the database, all occurrences of the RT are updated.
func UpdateRefreshToken(tx *sqlx.Tx, tokenID mtid.MTID, newRT, jwt string) error {
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

// ReencryptEncryptionKey re-encrypts the encryption key for a mytoken. This is needed when the mytoken changes, e.g. on token rotation
func ReencryptEncryptionKey(tx *sqlx.Tx, tokenID mtid.MTID, oldJWT, newJWT string) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		keyID, err := getEncryptionKeyID(tx, tokenID)
		if err != nil {
			return err
		}
		var encryptedKey string
		if err = tx.Get(&encryptedKey, `SELECT encryption_key FROM EncryptionKeys WHERE id=?`, keyID); err != nil {
			return err
		}
		key, err := cryptUtils.AES256Decrypt(encryptedKey, oldJWT)
		if err != nil {
			return err
		}
		updatedKey, err := cryptUtils.AES256Encrypt(key, newJWT)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`UPDATE EncryptionKeys SET encryption_key=? WHERE id=?`, updatedKey, keyID)
		return err
	})
}

// GetEncryptionKey returns the encryption key and the rtid for a mytoken
func GetEncryptionKey(tx *sqlx.Tx, tokenID mtid.MTID, jwt string) (key []byte, rtID uint64, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		var res struct {
			EncryptedKey encryptionKey `db:"encryption_key"`
			ID           uint64        `db:"rt_id"`
		}
		if err = tx.Get(&res, `SELECT encryption_key, rt_id FROM MyTokens WHERE id=?`, tokenID); err != nil {
			return err
		}
		rtID = res.ID
		key, err = res.EncryptedKey.decrypt(jwt)
		return err
	})
	return
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

// GetRefreshToken returns the refresh token for a mytoken id
func GetRefreshToken(tx *sqlx.Tx, myid mtid.MTID, jwt string) (string, bool, error) {
	var rt rtStruct
	found, err := helper.ParseError(db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&rt, `SELECT refresh_token, encryption_key FROM MyTokens WHERE id=?`, myid)
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

// CountRTOccurrences counts how many Mytokens use the passed refresh token
func CountRTOccurrences(tx *sqlx.Tx, rtID uint64) (count int, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&count, `SELECT COUNT(1) FROM MTokens WHERE rt_id=?`, rtID)
	})
	return
}

// GetRTID returns the refresh token id for a mytoken
func GetRTID(tx *sqlx.Tx, myID mtid.MTID) (rtID uint64, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&rtID, `SELECT rt_id FROM MTokens WHERE id=?`, myID)
	})
	return
}

// getEncryptionKeyID returns the id of the encryption key used for encrypting the RT linked to this mytoken
func getEncryptionKeyID(tx *sqlx.Tx, myID mtid.MTID) (keyID uint64, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&keyID, `SELECT key_id FROM RT_EncryptionKeys WHERE MT_id=? AND rt_id=(SELECT rt_id FROM MTokens WHERE id=?)`, myID, myID)
	})
	return
}
