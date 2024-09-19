package database

import (
	"database/sql"
	"golang.org/x/net/context"
)

type DBRepository interface {
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Ping() error
}
