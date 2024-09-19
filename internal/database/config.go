package database

import (
	"errors"
	"fmt"
)

type DBConfig struct {
	Driver string `json:"driver" yaml:"driver"`
	URL    string `env-required:"true" json:"url" yaml:"url" env:"POSTGRES_CONN"`
	Host   string `env-required:"true" json:"host" yaml:"host" env:"POSTGRES_HOST"`
	Port   int    `env-required:"true" json:"port" yaml:"port" env:"POSTGRES_PORT"`
	User   string `env-required:"true" json:"user" yaml:"user" env:"POSTGRES_USERNAME"`
	Passwd string `env-required:"true" json:"passwd" yaml:"passwd" env:"POSTGRES_PASSWORD"`
	DBName string `env-required:"true" json:"DBName" yaml:"DBName" env:"POSTGRES_DATABASE"`
}

func (c *DBConfig) Validate() error {
	if c.Driver == "" {
		return errors.New("driver not specified")
	}

	switch c.Driver {
	case "postgres":

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
