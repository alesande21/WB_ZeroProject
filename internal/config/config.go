package config

import (
	"WB_ZeroProject/internal/database"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrorNotFoundConfig = errors.New("Config not found!")
)

type Config struct {
	LowercaseKeywords bool               `json:"lowercase_keywords" yaml:"lowercase_keywords"`
	Connection        *database.DBConfig `json:"connection" yaml:"connection"`
}

func newConfig() *Config {
	cfg := &Config{}
	cfg.LowercaseKeywords = false
	return cfg
}

func GetDefaultConfig() (*Config, error) {
	cfg := newConfig()
	err := cfg.loadConfigParam("config.yml")
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func GetConfig(filePath string) (*Config, error) {

	return nil, nil
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
		return fmt.Errorf("failed to read config, %w", err)
	}

	err = json.Unmarshal(buf, c.Connection)
	if err != nil {
		return fmt.Errorf("failed unmarshalling, %w", err)
	}

	err = c.validate()
	if err != nil {
		return fmt.Errorf("failed driver validation, %w", err)
	}

	return nil
}
