package transfercoderepo

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/utils/hashutils"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/utils"
	"github.com/oidc-mytoken/server/shared/utils/cryptutils"
)

// proxyToken holds information for proxy tokens, i.e. tokens that proxy another token, e.g. a short token
type proxyToken struct {
	id    string
	token string
	mtID  mtid.MTID

	encryptedJWT string
	decryptedJWT string
}

// newProxyToken creates a new proxyToken of the given length
func newProxyToken(size int) *proxyToken {
	token := utils.RandASCIIString(size)
	return createProxyToken(token)
}

func createProxyToken(token string) *proxyToken {
	id := hashutils.SHA512Str([]byte(token))
	return &proxyToken{
		id:    id,
		token: token,
	}
}

// parseProxyToken parses the proxy token string into a proxyToken
func parseProxyToken(token string) *proxyToken {
	var id string
	if token != "" {
		id = hashutils.SHA512Str([]byte(token))
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
		pt.id = hashutils.SHA512Str([]byte(pt.token))
	}
	return pt.id
}

// SetJWT sets the jwt for this proxyToken
func (pt *proxyToken) SetJWT(jwt string, mID mtid.MTID) (err error) {
	pt.mtID = mID
	pt.decryptedJWT = jwt
	pt.encryptedJWT, err = cryptutils.AES256Encrypt(jwt, pt.token)
	return
}

// JWT returns the decrypted jwt that is linked to this proxyToken
func (pt *proxyToken) JWT(rlog log.Ext1FieldLogger, tx *sqlx.Tx) (jwt string, valid bool, err error) {
	jwt = pt.decryptedJWT
	if jwt != "" {
		valid = true
		return
	}
	if pt.encryptedJWT == "" {
		if err = db.RunWithinTransaction(
			rlog, tx, func(tx *sqlx.Tx) error {
				var res struct {
					JWT  string    `db:"jwt"`
					MTID mtid.MTID `db:"MT_id"`
				}
				if err = tx.Get(&res, `CALL ProxyTokens_GetMT(?)`, pt.id); err != nil {
					return errors.WithStack(err)
				}
				pt.encryptedJWT = res.JWT
				pt.mtID = res.MTID
				return nil
			},
		); err != nil {
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
	jwt, err = cryptutils.AES256Decrypt(pt.encryptedJWT, pt.token)
	pt.decryptedJWT = jwt
	return
}

// Store stores the proxyToken
func (pt proxyToken) Store(rlog log.Ext1FieldLogger, tx *sqlx.Tx) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL ProxyTokens_Insert(?,?,?)`, pt.id, pt.encryptedJWT, pt.mtID)
			return errors.WithStack(err)
		},
	)
}

// Update updates the jwt of the proxyToken
func (pt proxyToken) Update(rlog log.Ext1FieldLogger, tx *sqlx.Tx) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL ProxyTokens_Update(?,?,?)`, pt.ID(), pt.encryptedJWT, pt.mtID)
			return errors.WithStack(err)
		},
	)
}

// Delete deletes the proxyToken from the database, it does not delete the linked Mytoken, the jwt should have been
// retrieved earlier and the Mytoken if desired be revoked separately
func (pt proxyToken) Delete(rlog log.Ext1FieldLogger, tx *sqlx.Tx) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL ProxyTokens_Delete(?)`, pt.id)
			return errors.WithStack(err)
		},
	)
}
