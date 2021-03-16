package mytoken

import (
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
)

// UsedMytoken is a type for a Mytoken that has been used, it additionally has information how often it has been used
type UsedMytoken struct {
	Mytoken
	Restrictions []restrictions.UsedRestriction `json:"restrictions,omitempty"`
}

func (mt *Mytoken) ToUsedMytoken(tx *sqlx.Tx) (*UsedMytoken, error) {
	umt := &UsedMytoken{
		Mytoken: *mt,
	}
	var err error
	umt.Restrictions, err = mt.Restrictions.ToUsedRestrictions(tx, mt.ID)
	return umt, err
}
