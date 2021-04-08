package api

type EventHistory []EventEntry

type EventEntry struct {
	Event          string `db:"event" json:"event"`
	Time           int64  `db:"time" json:"time"`
	Comment        string `db:"comment" json:"comment,omitempty"`
	ClientMetaData `json:",inline"`
}
