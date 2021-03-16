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

func (st *Mytoken) ToUsedMytoken(tx *sqlx.Tx) (*UsedMytoken, error) {
	ust := &UsedMytoken{
		Mytoken: *st,
	}
	var err error
	ust.Restrictions, err = st.Restrictions.ToUsedRestrictions(tx, st.ID)
	return ust, err
}
