package accesstokenrepo

// accessTokenAttribute holds database information about an access token attribute
type accessTokenAttribute struct {
	ATID   uint64 `db:"AT_id"`
	AttrID uint64 `db:"attribute_id"`
	Attr   string `db:"attribute"`
}
