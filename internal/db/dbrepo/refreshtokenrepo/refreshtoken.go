package refreshtokenrepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
)

// CountRTOccurrences counts how many Mytokens use the passed refresh token
func CountRTOccurrences(tx *sqlx.Tx, rtID uint64) (count int, err error) {
	err = db.RunWithinTransaction(
		tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&count, `CALL RT_CountLinks(?)`, rtID))
		},
	)
	return
}

// GetRTID returns the refresh token id for a mytoken
func GetRTID(tx *sqlx.Tx, myID mtid.MTID) (rtID uint64, err error) {
	err = db.RunWithinTransaction(
		tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Get(&rtID, `CALL MTokens_GetRTID(?)`, myID))
		},
	)
	return
}
