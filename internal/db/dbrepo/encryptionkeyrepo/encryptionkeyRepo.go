package encryptionkeyrepo

import (
	"encoding/base64"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/utils/cryptUtils"
)

// ReencryptEncryptionKey re-encrypts the encryption key for a mytoken. This is needed when the mytoken changes, e.g. on
// token rotation
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

// DeleteEncryptionKey deletes the encryption key for a mytoken.
func DeleteEncryptionKey(tx *sqlx.Tx, tokenID mtid.MTID) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		keyID, err := getEncryptionKeyID(tx, tokenID)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(`DELETE FROM EncryptionKeys WHERE id=?`, keyID); err != nil {
			return err
		}
		return nil
	})
}

// GetEncryptionKey returns the encryption key and the rtid for a mytoken
func GetEncryptionKey(tx *sqlx.Tx, tokenID mtid.MTID, jwt string) (key []byte, rtID uint64, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		var res struct {
			EncryptedKey EncryptionKey `db:"encryption_key"`
			ID           uint64        `db:"rt_id"`
		}
		if err = tx.Get(&res, `SELECT encryption_key, rt_id FROM MyTokens WHERE id=?`, tokenID); err != nil {
			return err
		}
		rtID = res.ID
		key, err = res.EncryptedKey.Decrypt(jwt)
		return err
	})
	return
}

// EncryptionKey is a type for the encryption key stored in the db
type EncryptionKey string

// Decrypt returns the decrypted encryption key
func (k EncryptionKey) Decrypt(jwt string) ([]byte, error) {
	decryptedKey, err := cryptUtils.AES256Decrypt(string(k), jwt)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(decryptedKey)
}

// getEncryptionKeyID returns the id of the encryption key used for encrypting the RT linked to this mytoken
func getEncryptionKeyID(tx *sqlx.Tx, myID mtid.MTID) (keyID uint64, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&keyID,
			`SELECT key_id FROM RT_EncryptionKeys WHERE MT_id=? AND rt_id=(SELECT rt_id FROM MTokens WHERE id=?)`,
			myID, myID)
	})
	return
}
