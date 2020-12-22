package shorttokenrepo

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/zachmann/mytoken/internal/config"
	"github.com/zachmann/mytoken/internal/db"
	"github.com/zachmann/mytoken/internal/utils"
	"github.com/zachmann/mytoken/internal/utils/cryptUtils"
)

const (
	idLen          = 16
	minPasswordLen = 16
)

// ShortToken holds database information of a short token
type ShortToken struct {
	ID                 string `db:"short_token_id"`
	encryptionPassword string

	EncryptedJWT string `db:"jwt"`
	decryptedJWT string
}

func (st *ShortToken) String() string {
	return st.ID + st.encryptionPassword
}

// NewShortToken creates a new short token from the given jwt of a normal SuperToken
func NewShortToken(jwt string) (*ShortToken, error) {
	passwordLen := utils.MinInt(config.Get().Features.ShortTokens.Len-idLen, minPasswordLen)
	shortToken := &ShortToken{
		ID:                 utils.RandASCIIString(idLen),
		encryptionPassword: utils.RandASCIIString(passwordLen),
		decryptedJWT:       jwt,
	}
	err := shortToken.encrypt()
	return shortToken, err
}

// ParseShortToken creates a new short token from a short token string
func ParseShortToken(tx *sqlx.Tx, token string) (*ShortToken, bool, error) {
	if len(token) < idLen {
		return nil, false, nil
	}
	shortToken := &ShortToken{
		ID:                 token[:idLen],
		encryptionPassword: token[idLen:],
	}
	found := true
	err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return tx.Get(&shortToken.EncryptedJWT, `SELECT jwt FROM ShortSuperTokens WHERE short_token_id=?`, shortToken.ID)
	})
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		found = false
		err = nil
	}
	return shortToken, found, err
}

func (st *ShortToken) encrypt() (err error) {
	st.EncryptedJWT, err = cryptUtils.AES256Encrypt(st.decryptedJWT, st.encryptionPassword)
	return
}

// JWT returns the decrypted jwt of the long SuperToken
func (st *ShortToken) JWT() (jwt string, err error) {
	jwt = st.decryptedJWT
	if len(jwt) > 0 {
		return
	}
	jwt, err = cryptUtils.AES256Decrypt(st.EncryptedJWT, st.encryptionPassword)
	st.decryptedJWT = jwt
	return
}

// Store stores the ShortToken
func (st *ShortToken) Store(tx *sqlx.Tx) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.NamedExec(`INSERT INTO ShortSuperTokens (short_token_id, jwt) VALUES (:short_token_id, :jwt)`, st)
		return err
	})
}

// Delete deletes the short token from the database, it does not delete the linked SuperToken, the jwt should have been retrieved earlier and the SuperToken should be revoked separately
func (st *ShortToken) Delete(tx *sqlx.Tx) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM ShortSuperTokens WHERE short_token_id=?`, st.ID)
		return err
	})
}
