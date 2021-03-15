package transfercoderepo

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
	"github.com/oidc-mytoken/server/shared/supertoken/pkg/stid"
	"github.com/oidc-mytoken/server/shared/utils"
	"github.com/oidc-mytoken/server/shared/utils/cryptUtils"
)

// proxyToken holds information for proxy tokens, i.e. tokens that proxy another token, e.g. a short token
type proxyToken struct {
	id    string
	token string
	mtID  stid.STID

	encryptedJWT string
	decryptedJWT string
}

// newProxyToken creates a new proxyToken of the given length
func newProxyToken(size int) *proxyToken {
	token := utils.RandASCIIString(size)
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
	if token != "" {
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
	if pt.id == "" {
		pt.id = hashUtils.SHA512Str([]byte(pt.token))
	}
	return pt.id
}

// SetJWT sets the jwt for this proxyToken
func (pt *proxyToken) SetJWT(jwt string, mID stid.STID) (err error) {
	pt.mtID = mID
	pt.decryptedJWT = jwt
	pt.encryptedJWT, err = cryptUtils.AES256Encrypt(jwt, pt.token)
	return
}

// JWT returns the decrypted jwt that is linked to this proxyToken
func (pt *proxyToken) JWT(tx *sqlx.Tx) (jwt string, valid bool, err error) {
	jwt = pt.decryptedJWT
	if jwt != "" {
		valid = true
		return
	}
	if pt.encryptedJWT == "" {
		if err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
			var res struct {
				JWT  string    `db:"jwt"`
				MTID stid.STID `db:"MT_id"`
			}
			if err := tx.Get(&res, `SELECT jwt, MT_id FROM ProxyTokens WHERE id=?`, pt.id); err != nil {
				return err
			}
			pt.encryptedJWT = res.JWT
			pt.mtID = res.MTID
			return nil
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = nil
				return
			}
			return
		}
	}
	valid = true
	if pt.encryptedJWT == "" {
		return
	}
	jwt, err = cryptUtils.AES256Decrypt(pt.encryptedJWT, pt.token)
	pt.decryptedJWT = jwt
	return
}

// Store stores the proxyToken
func (pt proxyToken) Store(tx *sqlx.Tx) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO ProxyTokens (id, jwt, MT_id) VALUES (?,?,?)`, pt.id, pt.encryptedJWT, pt.mtID)
		return err
	})
}

// Update updates the jwt of the proxyToken
func (pt proxyToken) Update(tx *sqlx.Tx) error {
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`UPDATE ProxyTokens SET jwt=?, MT_id=? WHERE id=?`, pt.encryptedJWT, pt.mtID, pt.id)
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
