package profilerepo

import (
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/utils/utils/profile"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
)

// GetGroups returns the list of available groups
func GetGroups(rlog log.Ext1FieldLogger, tx *sqlx.Tx) (groups []string, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&groups, `CALL Profiles_GetGroups()`))
		},
	)
	return
}

type profileData struct {
	ID      uuid.UUID `db:"id"`
	Group   string    `db:"group"`
	Name    string    `db:"name"`
	Payload string    `db:"payload"`
}

// GetProfiles returns the profiles for a group
func GetProfiles(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string) (profiles []api.Profile, err error) {
	return profileGetAndParse(
		rlog, tx, group,
		func(group string, dbReader *dbProfileReader) (map[string]profileData, error) {
			return dbReader.readAllProfile(group)
		}, func(content []byte, parser *profile.Parser) (interface{}, error) {
			return parser.ParseProfile(content)
		},
	)
}

// GetRestrictionsTemplates returns the rotation templates for a group
func GetRestrictionsTemplates(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string) (profiles []api.Profile, err error) {
	return profileGetAndParse(
		rlog, tx, group,
		func(group string, dbReader *dbProfileReader) (map[string]profileData, error) {
			return dbReader.readAllRestrictionsTemplate(group)
		}, func(content []byte, parser *profile.Parser) (interface{}, error) {
			return parser.ParseRestrictionsTemplate(content)
		},
	)
}

// GetCapabilitiesTemplates returns the rotation templates for a group
func GetCapabilitiesTemplates(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string) (profiles []api.Profile, err error) {
	return profileGetAndParse(
		rlog, tx, group,
		func(group string, dbReader *dbProfileReader) (map[string]profileData, error) {
			return dbReader.readAllCapabilityTemplate(group)
		}, func(content []byte, parser *profile.Parser) (interface{}, error) {
			return parser.ParseCapabilityTemplate(content)
		},
	)
}

// GetRotationTemplates returns the rotation templates for a group
func GetRotationTemplates(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string) (profiles []api.Profile, err error) {
	return profileGetAndParse(
		rlog, tx, group,
		func(group string, dbReader *dbProfileReader) (map[string]profileData, error) {
			return dbReader.readAllRotationTemplate(group)
		}, func(content []byte, parser *profile.Parser) (interface{}, error) {
			return parser.ParseRotationTemplate(content)
		},
	)
}

func getProfiles(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string) (profiles []profileData, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&profiles, `CALL Profiles_GetProfiles(?)`, group))
		},
	)
	return
}
func getCapabilityTemplates(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string) (profiles []profileData, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&profiles, `CALL Profiles_GetCapabilities(?)`, group))
		},
	)
	return
}
func getRotationTemplates(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string) (profiles []profileData, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&profiles, `CALL Profiles_GetRotations(?)`, group))
		},
	)
	return
}
func getRestrictionsTemplates(rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string) (profiles []profileData, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			return errors.WithStack(tx.Select(&profiles, `CALL Profiles_GetRestrictions(?)`, group))
		},
	)
	return
}

func profileGetAndParse(
	rlog log.Ext1FieldLogger, tx *sqlx.Tx, group string,
	readAllFnc func(string, *dbProfileReader) (map[string]profileData, error),
	parseFnc func([]byte, *profile.Parser) (interface{}, error),
) (profiles []api.Profile, err error) {
	err = db.RunWithinTransaction(
		rlog, tx, func(tx *sqlx.Tx) error {
			dbReader := newDBProfileReader(rlog)
			p := profile.NewParser(dbReader)
			groupData, err := readAllFnc(group, dbReader)
			if err != nil {
				return err
			}
			for _, d := range groupData {
				rot, err := parseFnc([]byte(d.Payload), p)
				if err != nil {
					return err
				}
				parsedPayload, err := json.Marshal(rot)
				if err != nil {
					return errors.WithStack(err)
				}
				d.Payload = string(parsedPayload)
				profiles = append(
					profiles, api.Profile{
						ID:      d.ID.String(),
						Name:    d.Name,
						Payload: json.RawMessage(d.Payload),
					},
				)
			}
			return nil
		},
	)
	return
}
