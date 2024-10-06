package database

import (
	"fmt"
	log2 "github.com/sirupsen/logrus"
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
		return fmt.Errorf("-> c.Connection.Validate: драйвер не указан")
	}

	switch c.Driver {
	case "postgres":
		log2.Debugf("Драйвер %s найден", c.Driver)
	default:
		return fmt.Errorf("-> c.Connection.Validate: драйвер не найден")
	}
	return nil
}

func (c *DBConfig) GetConfigInfo() string {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Passwd, c.DBName)
	return psqlInfo
}
