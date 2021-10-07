package restrictions

import (
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
)

// UsedRestriction is a type for a restriction that has been used and additionally has information how often is has been
// used
type UsedRestriction struct {
	Restriction
	UsagesATDone    *int64 `json:"usages_AT_done,omitempty"`
	UsagesOtherDone *int64 `json:"usages_other_done,omitempty"`
}

// ToUsedRestrictions turns a Restrictions into a slice of UsedRestriction
func (r Restrictions) ToUsedRestrictions(tx *sqlx.Tx, id mtid.MTID) (ur []UsedRestriction, err error) {
	var u UsedRestriction
	for _, rr := range r {
		u, err = rr.ToUsedRestriction(tx, id)
		if err != nil {
			return
		}
		ur = append(ur, u)
	}
	return
}

// ToUsedRestriction turns a Restriction into an UsedRestriction
func (r Restriction) ToUsedRestriction(tx *sqlx.Tx, id mtid.MTID) (UsedRestriction, error) {
	ur := UsedRestriction{
		Restriction: r,
	}
	err := db.RunWithinTransaction(
		tx, func(tx *sqlx.Tx) error {
			at, err := r.getATUsageCounts(tx, id)
			if err != nil {
				return err
			}
			ur.UsagesATDone = at
			other, err := r.getOtherUsageCounts(tx, id)
			ur.UsagesOtherDone = other
			return err
		},
	)
	return ur, err
}
