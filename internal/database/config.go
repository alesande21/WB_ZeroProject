package database

import (
	"errors"
	"fmt"
)

type DBConfig struct {
	Driver string `json:"driver" yaml:"driver"`
	Host   string `json:"host" yaml:"host"`
	Port   int    `json:"port" yaml:"port"`
	User   string `json:"user" yaml:"user"`
	Passwd string `json:"passwd" yaml:"passwd"`
	DBName string `json:"DBName" yaml:"DBName"`
}

func (c *DBConfig) Validate() error {
	if c.Driver == "" {
		return errors.New("driver not specified")
	}

	switch c.Driver {
	case "postgresql":

	default:
		return errors.New("driver not specified")
	}
	return nil
}

func (c *DBConfig) GetConfigInfo() string {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Passwd, c.DBName)
	return psqlInfo
}
