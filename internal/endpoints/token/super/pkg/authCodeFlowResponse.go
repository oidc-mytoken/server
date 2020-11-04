package pkg

import (
	"encoding/json"
	"time"
)

type authCodeFlowResponse struct {
	AuthorizationURL   string `json:"authorization_url"`
	PollingCode        string `json:"polling_code,omitempty"`
	PollingCodeExpires int64  `json:"polling_code_expires,omitempty"`
}

type AuthCodeFlowResponse struct {
	AuthorizationURL   string
	PollingCode        string
	PollingCodeExpires time.Time
}

func (r AuthCodeFlowResponse) MarshalJSON() ([]byte, error) {
	rr := authCodeFlowResponse{
		AuthorizationURL: r.AuthorizationURL,
		PollingCode:      r.PollingCode,
	}
	if rr.PollingCode != "" {
		rr.PollingCodeExpires = r.PollingCodeExpires.Unix()
	}
	return json.Marshal(rr)
}

func (r AuthCodeFlowResponse) UnmarshalJSON(data []byte) error {
	rr := authCodeFlowResponse{}
	if err := json.Unmarshal(data, &rr); err != nil {
		return err
	}
	r.AuthorizationURL = rr.AuthorizationURL
	r.PollingCode = rr.PollingCode
	r.PollingCodeExpires = time.Unix(rr.PollingCodeExpires, 0)
	return nil
}
