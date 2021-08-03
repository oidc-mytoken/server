package refreshtokenrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/encryptionkeyrepo"
	helper "github.com/oidc-mytoken/server/internal/db/dbrepo/mytokenrepo/mytokenrepohelper"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/utils/cryptUtils"
)

// UpdateRefreshToken updates a refresh token in the database, all occurrences of the RT are updated.
func UpdateRefreshToken(tx *sqlx.Tx, tokenID mtid.MTID, newRT, jwt string) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		key, rtID, err := encryptionkeyrepo.GetEncryptionKey(tx, tokenID, jwt)
		if err != nil {
			return errors.WithStack(err)
		}
		updatedRT, err := cryptUtils.AESEncrypt(newRT, key)
		if err != nil {
			return errors.WithStack(err)
		}
		_, err = tx.Exec(`UPDATE RefreshTokens SET rt=? WHERE id=?`, updatedRT, rtID)
		return errors.WithStack(err)
	})
}

type rtStruct struct {
	RT  string                          `db:"refresh_token"`
	Key encryptionkeyrepo.EncryptionKey `db:"encryption_key"`
}

func (rt rtStruct) decrypt(jwt string) (string, error) {
	key, err := rt.Key.Decrypt(jwt)
	if err != nil {
		return "", err
	}
	return cryptUtils.AESDecrypt(rt.RT, key)
}

// GetRefreshToken returns the refresh token for a mytoken id
func GetRefreshToken(tx *sqlx.Tx, myid mtid.MTID, jwt string) (string, bool, error) {
	var rt rtStruct
	found, err := helper.ParseError(db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return errors.WithStack(tx.Get(&rt, `SELECT refresh_token, encryption_key FROM MyTokens WHERE id=?`, myid))
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
		return errors.WithStack(err)
	})
}

// CountRTOccurrences counts how many Mytokens use the passed refresh token
func CountRTOccurrences(tx *sqlx.Tx, rtID uint64) (count int, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return errors.WithStack(tx.Get(&count, `SELECT COUNT(1) FROM MTokens WHERE rt_id=?`, rtID))
	})
	return
}

// GetRTID returns the refresh token id for a mytoken
func GetRTID(tx *sqlx.Tx, myID mtid.MTID) (rtID uint64, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return errors.WithStack(tx.Get(&rtID, `SELECT rt_id FROM MTokens WHERE id=?`, myID))
	})
	return
}
