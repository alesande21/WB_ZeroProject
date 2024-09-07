package database

import (
	"WB_ZeroProject/pkg/server/api"
	"database/sql"
	"fmt"
)

type postgresDBRepository struct {
	Conn *sql.DB
}

func (p *postgresDBRepository) Query(query string, args ...interface{}) (api.Orders, error) {
	return nil, nil
}

func (p *postgresDBRepository) QueryRow(query string, args ...interface{}) (*api.Order, error) {
	return nil, nil
}

func (p *postgresDBRepository) Exec(query string, args ...interface{}) error {
	return nil
}

func (p *postgresDBRepository) Ping() error {
	if p == nil {
		return nil
	}

	err := p.Conn.Ping()
	if err != nil {
		return fmt.Errorf("problems connecting to the database: %s", err.Error())
	}

	return nil
}

/*
func (p *postgresDBRepository) Query(query string, args ...interface{}) (api.Commands, error) {
	pErr := p.Conn.Ping()
	if pErr != nil {
		log.Println("Problems connecting to the database!")
		return nil, pErr
	}

	var foundCommand api.Command
	var commands []api.Command

	rows, err := p.Conn.Query(query, args...)

	if err != nil {
		log.Println("Ошибка при поиске команды!", err)
		return commands, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&foundCommand.Id, &foundCommand.BodyScript, &foundCommand.ResultRunScript, &foundCommand.Status)
		if err != nil {
			log.Println("Ошибка при считывании строк:", err)
			return commands, err
		}
		commands = append(commands, foundCommand)
	}
	return commands, nil
}

func (p *postgresDBRepository) QueryRow(query string, args ...interface{}) (*api.Command, error) {
	pErr := p.Conn.Ping()
	if pErr != nil {
		log.Println("Problems connecting to the database!")
		return nil, pErr
	}

	var command api.Command
	err := p.Conn.QueryRow(query, args...).Scan(&command.Id, &command.BodyScript, &command.ResultRunScript, &command.Status)
	if err != nil {
		//log.Println("Failed to insert row:", err)
		return nil, err
	}

	return &command, nil
}

func (p *postgresDBRepository) Exec(query string, args ...interface{}) error {
	pErr := p.Conn.Ping()
	if pErr != nil {
		log.Println("Problems connecting to the database!")
		return pErr
	}

	_, err := p.Conn.Exec(query, args...)
	return err
}
*/
