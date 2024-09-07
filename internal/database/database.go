package database

import "WB_ZeroProject/pkg/server/api"

type DBRepository interface {
	Query(query string, args ...interface{}) (api.Orders, error)    //  (api.Commands, error)
	QueryRow(query string, args ...interface{}) (*api.Order, error) // (*api.Command, error)
	Exec(query string, args ...interface{}) error
	Ping() error
}
