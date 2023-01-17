package profiled

import (
	"github.com/oidc-mytoken/api/v0"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/db/profilerepo"
	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/mytoken/restrictions"
)

// GeneralMytokenRequest extends the api.GeneralMytokenRequest with profile unmarshalling
type GeneralMytokenRequest struct {
	api.GeneralMytokenRequest
	Restrictions Restrictions       `json:"restrictions,omitempty"`
	Capabilities Capabilities       `json:"capabilities,omitempty"`
	Rotation     *Rotation          `json:"rotation,omitempty"`
	GrantType    model.GrantType    `json:"grant_type"`
	ResponseType model.ResponseType `json:"response_type"`
}

// NewGeneralMytokenRequest creates a GeneralMytokenRequest with default values
func NewGeneralMytokenRequest() *GeneralMytokenRequest {
	return &GeneralMytokenRequest{
		GeneralMytokenRequest: api.GeneralMytokenRequest{
			ResponseType: api.ResponseTypeToken,
		},
		ResponseType: model.ResponseTypeToken,
		GrantType:    -1,
	}
}

// UnmarshalJSON implements the json.Marshaler interface
func (r *GeneralMytokenRequest) UnmarshalJSON(bytes []byte) error {
	parser := profilerepo.NewDBProfileParser(log.StandardLogger())
	p, err := parser.ParseProfile(bytes)
	if err != nil {
		return err
	}
	r.GeneralMytokenRequest = p
	r.Restrictions.Restrictions = restrictions.NewRestrictionsFromAPI(p.Restrictions)
	r.Capabilities.Capabilities = p.Capabilities
	if p.Rotation != nil {
		r.Rotation = &Rotation{
			Rotation: *p.Rotation,
		}
	}
	r.ResponseType = model.NewResponseType(p.ResponseType)
	r.GrantType = model.NewGrantType(p.GrantType)
	return nil
}
