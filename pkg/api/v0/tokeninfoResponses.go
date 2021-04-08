package api

type TokeninfoIntrospectResponse struct {
	Valid bool        `json:"valid"`
	Token UsedMytoken `json:"token"`
}

type TokeninfoHistoryResponse struct {
	EventHistory EventHistory `json:"events"`
}

type TokeninfoTreeResponse struct {
	Tokens MytokenEntryTree `json:"mytokens"`
}
type TokeninfoListResponse struct {
	Tokens []MytokenEntryTree `json:"mytokens"`
}
