package database

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type postgresDBRepository struct {
	// TODO: вместо соединения поставить функцию
	//Conn *sql.DB
	getConnection func() *sql.DB
}

func (p *postgresDBRepository) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return p.getConnection().QueryContext(ctx, query, args...)
}

func (p *postgresDBRepository) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return p.getConnection().QueryRowContext(ctx, query, args...)
}

func (p *postgresDBRepository) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return p.getConnection().ExecContext(ctx, query, args...)
}

func (p *postgresDBRepository) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return p.getConnection().BeginTx(ctx, opts)
}

func (p *postgresDBRepository) Ping() error {
	if p == nil {
		return fmt.Errorf("репозиторий не инициализирован")
	}

	err := p.getConnection().Ping()
	if err != nil {
		return fmt.Errorf("проблема с поключением к базе данных: %s", err.Error())
	}

	return nil
}

func CreatePostgresRepository(newConnection func() *sql.DB) (DBRepository, error) {
	var rep DBRepository = &postgresDBRepository{getConnection: newConnection}
	return rep, nil
}

/*
func CreatePostgresRepository(db *sql.DB, connChan chan *sql.DB) (DBRepository, error) {
	var rep DBRepository = &postgresDBRepository{Conn: db}
	go func() {
		for newConn := range connChan {
			log.Println("Обновление подключения к базе данных в репозитории...")
			rep.UpdateConn(newConn)
		}
	}()
	return rep, nil
}

func CreatePostgresRepository(db *sql.DB) (DBRepository, error) {
	var rep DBRepository = &postgresDBRepository{Conn: db}
	return rep, nil
}

func (p *postgresDBRepository) UpdateConn(updateConn *sql.DB) {
	p.Conn = updateConn
}
*/
