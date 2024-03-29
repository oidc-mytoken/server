package sshrepo

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	mytoken "github.com/oidc-mytoken/server/internal/mytoken/pkg"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/internal/utils/cryptutils"
)

// GetSSHInfo returns the SSHInfo stored in the database for the passed key and user hashes.
func GetSSHInfo(rlog log.Ext1FieldLogger, tx *sqlx.Tx, keyFP, userHash string) (info SSHInfo, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&info, `CALL SSHInfo_Get(?,?)`, keyFP, userHash))
		},
	)
	return
}

// GetAllSSHInfo returns the SSHInfo for all ssh keys for a given user
func GetAllSSHInfo(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myid mtid.MTID) (info []api.SSHKeyInfo, err error) {
	var dbInfo []SSHInfo
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&dbInfo, `CALL SSHInfo_GetAll(?)`, myid))
		},
	)
	if err != nil {
		return
	}
	for _, i := range dbInfo {
		apiI := api.SSHKeyInfo{
			Name:              i.Name.String,
			SSHKeyFingerprint: i.KeyFingerprint,
			Created:           i.Created.Unix(),
		}
		if i.LastUsed.Valid {
			apiI.LastUsed = utils.NewInt64(i.LastUsed.Time.Unix())
		}
		info = append(info, apiI)
	}
	return
}

// SSHInfo is a type holding the information stored in the database related to an ssh key
type SSHInfo struct {
	KeyID          string        `db:"key_id"`
	Name           db.NullString `db:"name"`
	KeyFingerprint string        `db:"ssh_key_fp"`
	UserHash       string        `db:"ssh_user_hash"`
	Created        time.Time     `db:"created"`
	LastUsed       sql.NullTime  `db:"last_used"`
	Enabled        db.BitBool    `db:"enabled"`
	EncryptedMT    string        `db:"MT_crypt"`
}

// Decrypt decrypts the encrypted mytoken linked to this ssh key with the passed password
func (i SSHInfo) Decrypt(password string) (*mytoken.Mytoken, error) {
	decryptedMT, err := cryptutils.AES256Decrypt(i.EncryptedMT, password)
	if err != nil {
		return nil, err
	}
	return mytoken.ParseJWT(decryptedMT)
}

// Delete deletes an ssh key for the given user (given by the mytoken) from the database
func Delete(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myid mtid.MTID, keyHash string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL SSHInfo_Delete(?,?)`, myid, keyHash)
			return errors.WithStack(err)
		},
	)
}

// SSHInfoIn is a type for storing an ssh public key in the database
type SSHInfoIn struct {
	MTID           mtid.MTID     `db:"MT_id"`
	Name           db.NullString `db:"name"`
	KeyFingerprint string        `db:"ssh_key_fp"`
	UserHash       string        `db:"ssh_user_hash"`
	EncryptedMT    string        `db:"MT_crypt"`
}

// Insert inserts an ssh public key for the given user (given by the mytoken) into the database
func Insert(rlog log.Ext1FieldLogger, tx *sqlx.Tx, data SSHInfoIn) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(
				`CALL SSHInfo_Insert(?,?,?,?,?)`,
				data.MTID, data.KeyFingerprint, data.UserHash, data.Name, data.EncryptedMT,
			)
			return errors.WithStack(err)
		},
	)
}

// UsedKey marks that the passed ssh key was just used
func UsedKey(rlog log.Ext1FieldLogger, tx *sqlx.Tx, keyFP, userHash string) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL SSHInfo_UsedKey(?,?)`, keyFP, userHash)
			return errors.WithStack(err)
		},
	)
}
