package database

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

type postgresDBRepository struct {
	// TODO: вместо соединения поставить функцию
	Conn *sql.DB
}

func (p *postgresDBRepository) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return p.Conn.QueryContext(ctx, query, args...)
}

func (p *postgresDBRepository) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return p.Conn.QueryRowContext(ctx, query, args...)
}

func (p *postgresDBRepository) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return p.Conn.ExecContext(ctx, query, args...)
}

func (p *postgresDBRepository) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return p.Conn.BeginTx(ctx, opts)
}

func (p *postgresDBRepository) Ping() error {
	if p == nil {
		return nil
	}
	//p.Mutex.Lock()
	//defer p.Unlock()

	err := p.Conn.Ping()
	if err != nil {
		return fmt.Errorf("проблема с поключением к базе данных: %s", err.Error())
	}

	return nil
}

func (p *postgresDBRepository) UpdateConn(updateConn *sql.DB) {
	p.Conn = updateConn
}

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
