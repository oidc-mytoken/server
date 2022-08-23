package mytokenrepohelper

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/encryptionkeyrepo"
	"github.com/oidc-mytoken/server/internal/utils/hashUtils"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
)

// recursiveRevokeMT revokes the passed mytoken as well as all children
func recursiveRevokeMT(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id interface{}) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL MTokens_RevokeRec(?)`, id)
			return errors.WithStack(err)
		},
	)
}

// CheckTokenRevoked checks if a Mytoken has been revoked. If it is a rotating mytoken and auto_revoke is enabled for
// this token, it might get triggered.
func CheckTokenRevoked(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID, seqno uint64, rot *api.Rotation) (
	revoked bool, err error,
) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if rot == nil {
				revoked, err = checkTokenRevoked(rlog, tx, id, seqno)
				return err
			}
			if rot.Lifetime > 0 {
				revoked, err = checkRotatingTokenRevoked(rlog, tx, id, seqno, rot.Lifetime)
			} else {
				revoked, err = checkTokenRevoked(rlog, tx, id, seqno)
			}
			if err != nil {
				return err
			}
			if !revoked || !rot.AutoRevoke {
				return nil
			}
			// At this point we know, that the token is not valid, we now check if it is not valid because of the seqno.
			idFound, err := checkTokenID(rlog, tx, id)
			if err != nil || !idFound {
				return err
			}
			return RevokeMT(rlog, tx, id, true)
		},
	)
	return
}

func checkTokenRevoked(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID, seqno uint64) (bool, error) {
	var count int
	if err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&count, `CALL MTokens_Check(?,?)`, id, seqno))
		},
	); err != nil {
		return true, err
	}
	if count > 0 { // token was found as Mytoken
		return false, nil
	}
	return true, nil
}

func checkTokenID(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID) (bool, error) {
	var count int
	if err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&count, `CALL MTokens_CheckID(?)`, id))
		},
	); err != nil {
		return false, err
	}
	return count > 0, nil
}

func checkRotatingTokenRevoked(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID, seqno, rotationLifetime uint64) (
	bool, error,
) {
	var count int
	if err := db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&count, `CALL MTokens_CheckRotating(?,?,?)`, id, seqno, rotationLifetime))
		},
	); err != nil {
		return true, err
	}
	if count > 0 { // token was found as Mytoken
		return false, nil
	}
	return true, nil
}

// UpdateSeqNo updates the sequence number of a mytoken, i.e. it rotates the mytoken. Don't forget to update the
// encryption key
func UpdateSeqNo(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID, seqno uint64) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL MTokens_UpdateSeqNo(?,?)`, id, seqno)
			return errors.WithStack(err)
		},
	)
}

// revokeMT revokes the passed mytoken but no children
func revokeMT(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id interface{}) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if err := encryptionkeyrepo.DeleteEncryptionKey(rlog, tx, id); err != nil {
				return err
			}
			_, err := tx.Exec(`CALL MTokens_Delete(?)`, id)
			return errors.WithStack(err)
		},
	)
}

// RevokeMT revokes the passed mytoken and depending on the recursive parameter also its children
func RevokeMT(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id interface{}, recursive bool) error {
	if recursive {
		return recursiveRevokeMT(rlog, tx, id)
	} else {
		return revokeMT(rlog, tx, id)
	}
}

// RevocationIDHasParent checks if the token for a revocation id is a child of the (potential) parent mytoken
func RevocationIDHasParent(rlog log.Ext1FieldLogger, tx *sqlx.Tx, revocationID string, parent mtid.MTID) (
	isParent bool, err error,
) {
	var count int
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&count, `CALL MTokens_IsParentOf(?,?)`, parent, revocationID))
		},
	)
	isParent = count > 0
	return
}

// GetTokenUsagesAT returns how often a Mytoken was used with a specific restriction to obtain an access token
func GetTokenUsagesAT(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myID mtid.MTID, restrictionHash string) (
	usages *int64, err error,
) {
	var usageCount int64
	if err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&usageCount, `CALL TokenUsages_GetAT(?,?)`, myID, restrictionHash))
		},
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No usage entry -> was not used before -> usages=nil
			err = nil // This is fine
			return
		}
		return
	}
	usages = &usageCount
	return
}

// GetTokenUsagesOther returns how often a Mytoken was used with a specific restriction to do something else than
// obtaining an access token
func GetTokenUsagesOther(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myID mtid.MTID, restrictionHash string) (
	usages *int64, err error,
) {
	var usageCount int64
	if err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&usageCount, `CALL TokenUsages_GetOther(?,?)`, myID, restrictionHash))
		},
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No usage entry -> was not used before -> usages=nil
			err = nil // This is fine
			return
		}
		return
	}
	usages = &usageCount
	return
}

// IncreaseTokenUsageAT increases the usage count for obtaining ATs with a Mytoken and the given restriction
func IncreaseTokenUsageAT(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myID mtid.MTID, jsonRestriction []byte) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(
				`CALL TokenUsages_IncrAT(?,?,?)`,
				myID, jsonRestriction, hashUtils.SHA512Str(jsonRestriction),
			)
			return errors.WithStack(err)
		},
	)
}

// IncreaseTokenUsageOther increases the usage count for other usages with a Mytoken and the given restriction
func IncreaseTokenUsageOther(rlog log.Ext1FieldLogger, tx *sqlx.Tx, myID mtid.MTID, jsonRestriction []byte) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(
				`CALL TokenUsages_IncrOther(?,?,?)`,
				myID, jsonRestriction, hashUtils.SHA512Str(jsonRestriction),
			)
			return errors.WithStack(err)
		},
	)
}

// GetMTName returns the name of the mytoken
func GetMTName(rlog log.Ext1FieldLogger, tx *sqlx.Tx, id mtid.MTID) (name db.NullString, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&name, `CALL MTokens_GetName(?)`, id))
		},
	)
	return
}
