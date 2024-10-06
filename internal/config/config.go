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
		return nil, fmt.Errorf("-> cfg.loadEnvParam%w", err)
	}
	return cfg, nil
}

func GetConfigFromFile(filePath string) (*Config, error) {
	if filePath == "" {
		return nil, fmt.Errorf(": путь до конфига filePath не указан")
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
		return fmt.Errorf("-> os.Stat: файл по указаному пути не найден %s", filePath)
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("-> os.OpenFile: ошибка при открытии файла %s: %w", filePath, err)
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("-> io.ReadAll: ошибка при чтении файла %s: %w", filePath, err)
	}

	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return fmt.Errorf("-> yaml.Unmarshal: ошибка при кодировании файла: %w", err)
	}

	err = cleanenv.UpdateEnv(c)
	if err != nil {
		return fmt.Errorf("-> cleanenv.UpdateEnv: ошибка при обновлении параметроа из переменныз окружения%w", err)
	}

	err = c.validate()
	if err != nil {
		return fmt.Errorf("-> c.validate%w", err)
	}

	return nil
}

func (c *Config) loadEnvParam() error {
	var newConf database.DBConfig
	if err := cleanenv.ReadEnv(&newConf); err != nil {
		return fmt.Errorf("-> cleanenv.ReadEnv: ошибка загрузки env параметров конфига для подключения к бд: %w", err)
	}
	c.Connection = &newConf
	c.Connection.Driver = "postgres"
	return nil
}
