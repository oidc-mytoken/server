package dbModels

type AccessTokenAttribute struct {
	ATID   uint64 `db:"AT_id"`
	AttrID uint64 `db:"attribute_id"`
	Attr   string `db:"attribute"`
}
