package tagrepo

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils/ternary"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/mytoken/pkg/mtid"
)

// CreateTag creates a tag for the user linked to the mtID with the specified
// color (or auto-generated if empty)
func CreateTag(rlog log.Ext1FieldLogger, tx *sqlx.Tx, tag string, color string, mtID mtid.MTID) error {
	color = strings.TrimPrefix(color, "#")
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Tags_Create(?,?,?)`, mtID, tag, ternary.If(color != "", color, nil))
			return errors.WithStack(err)
		},
	)
}

// DeleteTag deletes a tag for the user linked to the mtID
func DeleteTag(rlog log.Ext1FieldLogger, tx *sqlx.Tx, tag string, mtID mtid.MTID) error {
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec(`CALL Tags_Delete(?,?)`, mtID, tag)
			return errors.WithStack(err)
		},
	)
}

// UpdateTag updates a tag for the user linked to the mtID
func UpdateTag(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx,
	oldTag string, tagInfo api.TagInfo, mtID mtid.MTID,
) error {
	color := strings.TrimPrefix(tagInfo.Color, "#")
	return db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			if color != "" {
				_, err := tx.Exec(
					`CALL Tags_UpdateColor(?,?,?)`, mtID, oldTag, color,
				)
				if err != nil {
					return errors.WithStack(err)
				}
			}
			if tagInfo.Tag != "" {
				_, err := tx.Exec(
					`CALL Tags_UpdateName(?,?,?)`, mtID, oldTag, tagInfo.Tag,
				)
				if err != nil {
					return errors.WithStack(err)
				}
			}
			return nil
		},
	)
}

// ListTags lists all tags of the user linked to the mtID
func ListTags(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx,
	mtID mtid.MTID,
) (tags []api.TagInfo, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&tags, `CALL Tags_List(?)`, mtID))
		},
	)
	return
}
