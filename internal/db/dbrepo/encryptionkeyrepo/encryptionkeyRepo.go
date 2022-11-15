package encryptionkeyrepo

import (
	"encoding/base64"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/utils/cryptutils"
)

// ReencryptEncryptionKey re-encrypts the encryption key for a mytoken. This is needed when the mytoken changes, e.g. on
// token rotation
func ReencryptEncryptionKey(rlog log.Ext1FieldLogger, tx *sqlx.Tx, tokenID mtid.MTID, oldJWT, newJWT string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			keyID, err := getEncryptionKeyID(rlog, tx, tokenID)
			if err != nil {
				return err
			}
			var encryptedKey string
			if err = errors.WithStack(tx.Get(&encryptedKey, `CALL EncryptionKeys_Get(?)`, keyID)); err != nil {
				return err
			}
			key, err := cryptutils.AES256Decrypt(encryptedKey, oldJWT)
			if err != nil {
				return err
			}
			updatedKey, err := cryptutils.AES256Encrypt(key, newJWT)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`CALL EncryptionKeys_Update(?,?)`, keyID, updatedKey)
			return errors.WithStack(err)
		},
	)
}

// DeleteEncryptionKey deletes the encryption key for a mytoken.
func DeleteEncryptionKey(rlog log.Ext1FieldLogger, tx *sqlx.Tx, tokenID interface{}) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			keyID, err := getEncryptionKeyID(rlog, tx, tokenID)
			if err != nil {
				return err
			}
			if _, err = tx.Exec(`CALL EncryptionKeys_Delete(?)`, keyID); err != nil {
				return errors.WithStack(err)
			}
			return nil
		},
	)
}

// GetEncryptionKey returns the encryption key and the rtid for a mytoken
func GetEncryptionKey(rlog log.Ext1FieldLogger, tx *sqlx.Tx, tokenID mtid.MTID, jwt string) (
	key []byte, rtID uint64, err error,
) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			var res RTCryptKeyDBRes
			if err = tx.Get(&res, `CALL EncryptionKeys_GetRTKeyForMT(?)`, tokenID); err != nil {
				return errors.WithStack(err)
			}
			rtID = res.RTID
			key, err = res.EncryptionKey.Decrypt(jwt)
			return err
		},
	)
	return
}

// EncryptionKey is a type for the encryption key stored in the db
type EncryptionKey string

// Decrypt returns the decrypted encryption key
func (k EncryptionKey) Decrypt(jwt string) ([]byte, error) {
	decryptedKey, err := cryptutils.AES256Decrypt(string(k), jwt)
	if err != nil {
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(decryptedKey)
	return data, errors.WithStack(err)
}

// RTCryptKeyDBRes is a struct holding the db result for the EncryptionKeys_GetRTKeyForMT procedure
type RTCryptKeyDBRes struct {
	KeyID         uint64        `db:"key_id"`
	EncryptionKey EncryptionKey `db:"encryption_key"`
	RTID          uint64        `db:"rt_id"`
	RT            string        `db:"refresh_token"`
}

// Decrypt returns the decrypted refresh token
func (res RTCryptKeyDBRes) Decrypt(jwt string) (string, error) {
	key, err := res.EncryptionKey.Decrypt(jwt)
	if err != nil {
		return "", err
	}
	return cryptutils.AESDecrypt(res.RT, key)
}

// getEncryptionKeyID returns the id of the encryption key used for encrypting the RT linked to this mytoken
func getEncryptionKeyID(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myID interface{}) (uint64, error) {
	var res RTCryptKeyDBRes
	err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&res, `CALL EncryptionKeys_GetRTKeyForMT(?)`, myID))
		},
	)
	return res.KeyID, err
}
