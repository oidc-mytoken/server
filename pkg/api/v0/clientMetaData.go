package api

// ClientMetaData hold information about the calling client
type ClientMetaData struct {
	IP        string `db:"ip" json:"ip,omitempty"`
	UserAgent string `db:"user_agent" json:"user_agent,omitempty"`
}
