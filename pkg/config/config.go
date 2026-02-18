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
	S3                *S3Config         `yaml:"s3_config"`
}

type S3Config struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	UseSSL    bool   `yaml:"use_ssl"`
	Bucket    string `yaml:"bucket"`
	Enabled   bool   `yaml:"enabled"`
	Region    string `yaml:"region"`
	Prefix    string `yaml:"prefix"`
}

type WarehouseConfig struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Enabled bool   `yaml:"enabled"`

	TableType string `yaml:"table_type"`

	StorageLocation string `yaml:"storage_location"`

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
	Properties  map[string]string `yaml:"properties,omitempty"`
}

func (w *WarehouseConfig) GetSchemaName(baseSchema string) string {

	schema := baseSchema

	if w.StorageLocation == "s3" {
		schema = schema + "_s3"
	}

	if w.TableType == "iceberg" {
		schema = schema + "_iceberg"
	}
	return schema
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

	if c.S3 != nil && c.S3.Enabled {
		if c.S3.Endpoint == "" {
			return fmt.Errorf("s3: endpoint не установлен")
		}

		if c.S3.AccessKey == "" {
			return fmt.Errorf("s3: access_key не установлен")
		}

		if c.S3.SecretKey == "" {
			return fmt.Errorf("s3: secret_key не установлен")
		}

		if c.S3.Bucket == "" {
			return fmt.Errorf("s3: bucket не установлен")
		}

		if c.S3.UseSSL && c.CertPath == "" {
			return fmt.Errorf("путь к сертификату не установлен")
		}
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
