package dbModels

type DBModel interface {
	Store() error
}
