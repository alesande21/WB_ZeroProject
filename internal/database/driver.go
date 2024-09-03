package database

import (
	"WB_ZeroProject/internal/colorAttribute"
	"database/sql"
	"fmt"
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

	db, err := sql.Open(cfg.Driver, psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("driver not found, %s", cfg.Driver)
	}

	log.Println(colorAttribute.ColorString(colorAttribute.FgYellow, "Успешное подключение к базе данных!"))

	return &DBConnection{Conn: db}, nil
}

func CreateRepository(db *sql.DB) (DBRepository, error) {
	var rep DBRepository = &postgresDBRepository{Conn: db}
	return rep, nil
}
