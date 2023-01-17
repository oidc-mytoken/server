package pkg

import (
	"database/sql/driver"
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/oidc-mytoken/api/v0"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/server/internal/model"
	"github.com/oidc-mytoken/server/internal/model/profiled"
)

// OIDCFlowRequest holds the request for an OIDC Flow request
type OIDCFlowRequest struct {
	profiled.GeneralMytokenRequest
	OIDCFlowAttrs
}

// OIDCFlowAttrs holds the additional attributes for OIDCFlowRequests
type OIDCFlowAttrs struct {
	OIDCFlow model.OIDCFlow `json:"oidc_flow"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (r *OIDCFlowRequest) UnmarshalJSON(data []byte) error {
	if err := errors.WithStack(json.Unmarshal(data, &r.GeneralMytokenRequest)); err != nil {
		return err
	}
	if len(r.Capabilities.Capabilities) == 0 {
		r.Capabilities.Capabilities = api.DefaultCapabilities
	}
	return errors.WithStack(json.Unmarshal(data, &r.OIDCFlowAttrs))
}

// MarshalJSON implements the json.Marshaler interface
func (r OIDCFlowRequest) MarshalJSON() ([]byte, error) {
	base, err := json.Marshal(r.GeneralMytokenRequest)
	if err != nil {
		return nil, err
	}
	attrs, err := json.Marshal(r.OIDCFlowAttrs)
	if err != nil {
		return nil, err
	}
	return jsonpatch.MergePatch(base, attrs)
}

// NewOIDCFlowRequest creates a new OIDCFlowRequest with default values where they can be omitted
func NewOIDCFlowRequest() *OIDCFlowRequest {
	return &OIDCFlowRequest{
		GeneralMytokenRequest: *profiled.NewGeneralMytokenRequest(),
		// clientType:            api.ClientTypeNative,
	}
}

// Scan implements the sql.Scanner interface
func (r *OIDCFlowRequest) Scan(src interface{}) error {
	v, ok := src.([]byte)
	if !ok {
		return errors.New("bad []byte type assertion")
	}
	return errors.WithStack(json.Unmarshal(v, r))
}

// Value implements the driver.Valuer interface
func (r OIDCFlowRequest) Value() (driver.Value, error) { // skipcq: CRT-P0003
	v, err := json.Marshal(r)
	return v, errors.WithStack(err)
}
