package transfercoderepo

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/zachmann/mytoken/internal/server/db"
	"github.com/zachmann/mytoken/internal/server/utils/hashUtils"
	"github.com/zachmann/mytoken/internal/utils"
	"github.com/zachmann/mytoken/internal/utils/cryptUtils"
)

// proxyToken holds information for proxy tokens, i.e. tokens that proxy another token, e.g. a short token
type proxyToken struct {
	id    string
	token string

	encryptedJWT string
	decryptedJWT string
}

// newProxyToken creates a new proxyToken of the given length
func newProxyToken(len int) *proxyToken {
	token := utils.RandASCIIString(len)
	return createProxyToken(token)
}

func createProxyToken(token string) *proxyToken {
	id := hashUtils.SHA512Str([]byte(token))
	return &proxyToken{
		id:    id,
		token: token,
	}
}

// parseProxyToken parses the proxy token string into a proxyToken
func parseProxyToken(token string) *proxyToken {
	var id string
	if len(token) > 0 {
		id = hashUtils.SHA512Str([]byte(token))
	}
	return &proxyToken{
		id:    id,
		token: token,
	}
}

func (pt proxyToken) String() string {
	return pt.Token()
}

// Token returns the token of this proxyToken
func (pt proxyToken) Token() string {
	return pt.token
}

// ID returns the id of this token
func (pt *proxyToken) ID() string {
	if len(pt.id) == 0 {
		pt.id = hashUtils.SHA512Str([]byte(pt.token))
	}
	return pt.id
}

// SetJWT sets the jwt for this proxyToken
func (pt *proxyToken) SetJWT(jwt string) (err error) {
	pt.decryptedJWT = jwt
	pt.encryptedJWT, err = cryptUtils.AES256Encrypt(jwt, pt.token)
	return
}

// JWT returns the decrypted jwt that is linked to this proxyToken
func (pt *proxyToken) JWT(tx *sqlx.Tx) (jwt string, valid bool, err error) {
	jwt = pt.decryptedJWT
	if len(jwt) > 0 {
		valid = true
		return
	}
	if len(pt.encryptedJWT) == 0 {
		if err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
			return tx.Get(&pt.encryptedJWT, `SELECT jwt FROM ProxyTokens WHERE id=?`, pt.id)
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = nil
				return
			}
			return
		}
	}
	valid = true
	if len(pt.encryptedJWT) == 0 {
		return
	}
	jwt, err = cryptUtils.AES256Decrypt(pt.encryptedJWT, pt.token)
	pt.decryptedJWT = jwt
	return
}

// Store stores the proxyToken
func (pt proxyToken) Store(tx *sqlx.Tx) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO ProxyTokens (id, jwt) VALUES (?,?)`, pt.id, pt.encryptedJWT)
		return err
	})
}

// Update updates the jwt of the proxyToken
func (pt proxyToken) Update(tx *sqlx.Tx) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`UPDATE ProxyTokens SET jwt=? WHERE id=?`, pt.encryptedJWT, pt.id)
		return err
	})
}

// Delete deletes the proxyToken from the database, it does not delete the linked SuperToken, the jwt should have been retrieved earlier and the SuperToken if desired be revoked separately
func (pt proxyToken) Delete(tx *sqlx.Tx) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM ProxyTokens WHERE id=?`, pt.id)
		return err
	})
}
