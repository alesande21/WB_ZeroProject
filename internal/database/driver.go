package database

import "database/sql"

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

func (db *DBConnection) Open(cfg *DBConfig) (*DBConnection, error) {

}
