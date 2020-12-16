package dbModels

// DBModel is an interface for database models
type DBModel interface {
	Store() error
}
