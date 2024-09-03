package database

type DBRepository interface {
	Query(query string, args ...interface{}) (error, error)    //  (api.Commands, error)
	QueryRow(query string, args ...interface{}) (error, error) // (*api.Command, error)
	Exec(query string, args ...interface{}) error
}
