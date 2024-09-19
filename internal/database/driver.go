package database

import (
	"AvitoProject/internal/colorAttribute"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

type DBConnection struct {
	Conn *sql.DB
}

func (db *DBConnection) Close() error {
	if db == nil {
		return nil
	}

	err := db.Conn.Close()
	if err != nil {
		return err
	}

	return nil
}

func (db *DBConnection) Ping() error {
	if db == nil {
		return nil
	}

	err := db.Conn.Ping()
	if err != nil {
		return fmt.Errorf("problems connecting to the database: %s", err.Error())
	}

	return nil
}

func Open(cfg *DBConfig) (*DBConnection, error) {
	psqlInfo := cfg.GetConfigInfo()

	log.Println(cfg.URL)
	log.Println(psqlInfo)
	db, err := sql.Open(cfg.Driver, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("driver not found, %s", cfg.Driver)
	}

	log.Println(colorAttribute.ColorString(colorAttribute.FgYellow, "Успешное подключение к базе данных!"))

	return &DBConnection{Conn: db}, nil
}

func (db *DBConnection) GetConn() (*sql.DB, error) {
	if db.Conn == nil {
		return nil, fmt.Errorf("драйвер подключения отсутсвует")
	}
	return db.Conn, nil
}

func (db *DBConnection) GetConn2() *sql.DB {
	return db.Conn
}
