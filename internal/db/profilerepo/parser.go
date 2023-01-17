package profilerepo

import (
	"github.com/jmoiron/sqlx"
	"github.com/oidc-mytoken/utils/utils/profile"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/utils"
)

type dbProfileReader struct {
	rlog             log.Ext1FieldLogger
	profileData      map[string]map[string]profileData
	restrictionsData map[string]map[string]profileData
	rotationData     map[string]map[string]profileData
	capabilityData   map[string]map[string]profileData
}

// ReadProfile implements the profile.TemplateReader interface
func (p dbProfileReader) ReadProfile(s string) ([]byte, error) {
	return read(p.rlog, s, p.profileData, getProfiles)
}

// ReadRestrictionsTemplate implements the profile.TemplateReader interface
func (p dbProfileReader) ReadRestrictionsTemplate(s string) ([]byte, error) {
	return read(p.rlog, s, p.restrictionsData, getRestrictionsTemplates)
}

// ReadRotationTemplate implements the profile.TemplateReader interface
func (p dbProfileReader) ReadRotationTemplate(s string) ([]byte, error) {
	return read(p.rlog, s, p.rotationData, getRotationTemplates)
}

// ReadCapabilityTemplate implements the profile.TemplateReader interface
func (p dbProfileReader) ReadCapabilityTemplate(s string) ([]byte, error) {
	return read(p.rlog, s, p.capabilityData, getCapabilityTemplates)
}

func (p dbProfileReader) readAllProfile(group string) (map[string]profileData, error) {
	return readGroupData(p.rlog, group, p.profileData, getProfiles)
}

func (p dbProfileReader) readAllRestrictionsTemplate(group string) (map[string]profileData, error) {
	return readGroupData(p.rlog, group, p.restrictionsData, getRestrictionsTemplates)
}

func (p dbProfileReader) readAllRotationTemplate(group string) (map[string]profileData, error) {
	return readGroupData(p.rlog, group, p.rotationData, getRotationTemplates)
}

func (p dbProfileReader) readAllCapabilityTemplate(group string) (map[string]profileData, error) {
	return readGroupData(p.rlog, group, p.capabilityData, getCapabilityTemplates)
}

func newDBProfileReader(rlog log.Ext1FieldLogger) *dbProfileReader {
	return &dbProfileReader{
		rlog:             rlog,
		profileData:      make(map[string]map[string]profileData),
		capabilityData:   make(map[string]map[string]profileData),
		restrictionsData: make(map[string]map[string]profileData),
		rotationData:     make(map[string]map[string]profileData),
	}
}

// NewDBProfileParser creates a new profile.ProfileParser that can read profiles from the db
func NewDBProfileParser(rlog log.Ext1FieldLogger) *profile.ProfileParser {
	return profile.NewProfileParser(newDBProfileReader(rlog))
}

func read(
	rlog log.Ext1FieldLogger,
	nameIncludingGroup string,
	dataSrc map[string]map[string]profileData, dbReader func(
		log.Ext1FieldLogger, *sqlx.Tx, string,
	) ([]profileData, error),
) (dataRead []byte, err error) {
	split := utils.SplitIgnoreEmpty(nameIncludingGroup, "/")
	if len(split) > 2 {
		return nil, errors.New("malformed include name")
	}
	name := split[0]
	group := "_"
	if len(split) == 2 {
		name = split[1]
		group = split[0]
	}
	groupData, err := readGroupData(rlog, group, dataSrc, dbReader)
	if err != nil {
		return nil, err
	}

	data, dataFound := groupData[name]
	if !dataFound {
		return nil, errors.Errorf("unknown include name: '%s'", nameIncludingGroup)
	}
	dataRead = []byte(data.Payload)
	return
}

func readGroupData(
	rlog log.Ext1FieldLogger,
	group string,
	dataSrc map[string]map[string]profileData, dbReader func(
		log.Ext1FieldLogger, *sqlx.Tx, string,
	) ([]profileData, error),
) (
	groupData map[string]profileData,
	err error,
) {
	if group == "" {
		group = "_"
	}
	var found bool
	groupData, found = dataSrc[group]
	if found {
		return
	}
	data, e := dbReader(rlog, nil, group)
	foundDB, dbErr := db.ParseError(e)
	if dbErr != nil {
		return nil, dbErr
	}
	if !foundDB {
		return nil, errors.New("unknown include name")
	}
	groupData = make(map[string]profileData)
	for _, d := range data {
		groupData[d.Name] = d
	}
	dataSrc[group] = groupData
	return
}
