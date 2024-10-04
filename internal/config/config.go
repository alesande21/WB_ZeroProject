package config

import (
	"WB_ZeroProject/internal/database"
	"errors"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/invopop/yaml"
	"io"
	"os"
)

var (
	ErrorNotFoundConfig = errors.New("config not found")
)

type Config struct {
	LowercaseKeywords bool               `json:"lowercaseKeywords" yaml:"lowercaseKeywords"`
	Connection        *database.DBConfig `json:"connection" yaml:"connection"`
}

func newConfig() *Config {
	cfg := &Config{}
	cfg.LowercaseKeywords = false
	return cfg
}

func GetDefaultConfig() (*Config, error) {
	cfg := newConfig()
	//err := cfg.loadConfigParam("src/internal/config/config.yml")
	err := cfg.loadEnvParam()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func GetConfig(filePath string) (*Config, error) {
	if filePath == "" {
		return nil, fmt.Errorf("путь до конфига не указан")
	}

	cfg := newConfig()
	err := cfg.loadConfigParam(filePath)
	//err := cfg.loadEnvParam()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) GetDBsConfig() *database.DBConfig {
	return c.Connection
}

func (c *Config) validate() error {
	if c.Connection == nil {
		return c.Connection.Validate()
	}
	return nil
}

func (c *Config) loadConfigParam(filePath string) error {
	_, err := os.Stat(filePath)
	if !(err == nil || !os.IsNotExist(err)) {
		return ErrorNotFoundConfig
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to read config, %w", err)
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read config, %w, %s", err, string(buf))
	}

	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return fmt.Errorf("failed unmarshalling, %w", err)
	}

	err = cleanenv.UpdateEnv(c)
	if err != nil {
		return fmt.Errorf("error updating env: %w", err)
	}

	err = c.validate()
	if err != nil {
		return fmt.Errorf("failed driver validation, %w", err)
	}

	return nil
}

func (c *Config) loadEnvParam() error {
	var newConf database.DBConfig
	if err := cleanenv.ReadEnv(&newConf); err != nil {
		return err
	}
	c.Connection = &newConf
	c.Connection.Driver = "postgres"
	return nil
}

type ConfigKafka struct {
	URL   string `env-required:"true" json:"url" yaml:"url" env:"KAFKA_CONN"`
	Host  string `env-required:"true" json:"host" yaml:"host" env:"KAFKA_HOST"`
	Port  int    `env-required:"true" json:"port" yaml:"port" env:"KAFKA_PORT"`
	Topic string `env-required:"true" json:"topic" yaml:"topic" env:"KAFKA_TOPIC"`
}

func GetConfigProducer() (*ConfigKafka, error) {
	var newConf ConfigKafka
	if err := cleanenv.ReadEnv(&newConf); err != nil {
		return nil, err
	}

	return &newConf, nil
}
