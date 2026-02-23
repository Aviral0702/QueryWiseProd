package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Database   DatabaseConfig   `yaml:"database"`
	Backend    BackendConfig    `yaml:"backend"`
	Collection CollectionConfig `yaml:"collection"`
	Agent      AgentConfig      `yaml:"agent"`
	Logging    LoggingConfig    `yaml:"logging"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type BackendConfig struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

type CollectionConfig struct {
	IntervalSeconds int `yaml:"interval_seconds"`
	BatchSize       int `yaml:"batch_size"`
	TimeoutSeconds  int `yaml:"timeout_seconds"`
}

type AgentConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
	Version     string `yaml:"version"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	// Defaults
	if c.Database.Port == 0 {
		c.Database.Port = 5432
	}
	if c.Collection.IntervalSeconds == 0 {
		c.Collection.IntervalSeconds = 60
	}
	if c.Collection.BatchSize == 0 {
		c.Collection.BatchSize = 100
	}
	if c.Agent.Name == "" {
		c.Agent.Name = "default-agent"
	}
	return &c, nil
}
