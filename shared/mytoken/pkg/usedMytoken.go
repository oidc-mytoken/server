package mytoken

import (
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/shared/mytoken/restrictions"
)

// UsedMytoken is a type for a Mytoken that has been used, it additionally has information how often it has been used
type UsedMytoken struct {
	Mytoken
	Restrictions []restrictions.UsedRestriction `json:"restrictions,omitempty"`
}

// ToUsedMytoken turns a Mytoken into a UsedMytoken by adding information about its usages
func (mt *Mytoken) ToUsedMytoken(rlog log.Ext1FieldLogger, tx *sqlx.Tx) (*UsedMytoken, error) {
	umt := &UsedMytoken{
		Mytoken: *mt,
	}
	var err error
	umt.Restrictions, err = mt.Restrictions.ToUsedRestrictions(rlog, tx, mt.ID)
	return umt, err
}
