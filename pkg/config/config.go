package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Warehouses        []WarehouseConfig `yaml:"warehouses"`
	Schema            string            `yaml:"schema"`
	QueriesPath       string            `yaml:"queries_path"`
	ResultsPath       string            `yaml:"results_path"`
	Timeout           string            `yaml:"timeout"`
	ConnectionTimeout string            `yaml:"connection_timeout"`
	CertPath          string            `yaml:"cert_path"`
	Runs              int               `yaml:"runs"`
	Concurrency       int               `yaml:"concurrency"`
	ConnectionRetries int               `yaml:"connection_retries"`
	RetryDelay        string            `yaml:"retry_delay"`
}

type WarehouseConfig struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Enabled bool   `yaml:"enabled"`

	TableType string `yaml:"table_type"`

	//Параметры подключения
	Connection ConnectionConfig `yaml:"connection"`
}

type ConnectionConfig struct {
	Host     string `yaml:"host,omitempty"`
	Port     string `yaml:"port,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Database string `yaml:"database,omitempty"`
	UseTLS   bool   `yaml:"use_tls,omitempty"`

	// Hive/Spark
	ZKQuorum    string            `yaml:"zk_quorum,omitempty"`
	ZKNamespace string            `yaml:"zk_namespace,omitempty"`
	HiveConfig  map[string]string `yaml:"hive_config,omitempty"`
}

func (w *WarehouseConfig) GetSchemaName(baseSchema string) string {
	if w.TableType == "iceberg" {
		return baseSchema + "_iceberg"
	}
	return baseSchema
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения конфигурационного файла: %w", err)

	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга конфигурации: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if len(c.Warehouses) == 0 {
		return fmt.Errorf("нет хранилищ данных")
	}

	if c.Schema == "" {
		return fmt.Errorf("схема не установлена")
	}

	if c.QueriesPath == "" {
		return fmt.Errorf("queries_path не установлен")
	}

	if c.ResultsPath == "" {
		return fmt.Errorf("results_path не установлен")
	}

	if c.Runs < 1 {
		c.Runs = 1
	}

	if c.Concurrency < 1 {
		c.Concurrency = 1
	}

	if c.ConnectionRetries < 1 {
		c.ConnectionRetries = 3
	}

	if c.RetryDelay == "" {
		c.RetryDelay = "5s"
	}

	return nil
}
