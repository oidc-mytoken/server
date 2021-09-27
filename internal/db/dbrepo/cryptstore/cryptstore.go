package cryptstore

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/encryptionkeyrepo"
	helper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/utils/cryptUtils"
)

// DeleteCrypted deletes an entry from the CryptStore
func DeleteCrypted(tx *sqlx.Tx, cryptID uint64) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`CALL CryptStore_Delete(?)`, cryptID)
		return errors.WithStack(err)
	})
}

// UpdateRefreshToken updates a refresh token in the database, all occurrences of the RT are updated.
func UpdateRefreshToken(tx *sqlx.Tx, tokenID mtid.MTID, newRT, jwt string) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		key, rtID, err := encryptionkeyrepo.GetEncryptionKey(tx, tokenID, jwt)
		if err != nil {
			return err
		}
		updatedRT, err := cryptUtils.AESEncrypt(newRT, key)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`CALL CryptStore_Update(?, ?)`, rtID, updatedRT)
		return errors.WithStack(err)
	})
}

// GetRefreshToken returns the refresh token for a mytoken id
func GetRefreshToken(tx *sqlx.Tx, myid mtid.MTID, jwt string) (string, bool, error) {
	var rt encryptionkeyrepo.RTCryptKeyDBRes
	found, err := helper.ParseError(db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return errors.WithStack(tx.Get(&rt, `CALL EncryptionKeys_GetRTKeyForMT(?)`, myid))
	}))
	if !found {
		return "", found, err
	}
	plainRT, err := rt.Decrypt(jwt)
	return plainRT, true, err
}
