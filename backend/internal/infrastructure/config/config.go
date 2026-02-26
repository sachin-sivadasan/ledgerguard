package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Firebase FirebaseConfig `yaml:"firebase"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
}

type FirebaseConfig struct {
	CredentialsFile string `yaml:"credentials_file"`
}

// Load loads configuration from file and environment variables.
// Priority: defaults < config file < environment variables
func Load(configPath string) (*Config, error) {
	// Start with defaults
	cfg := &Config{
		Server: ServerConfig{
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host:    "localhost",
			Port:    "5432",
			User:    "postgres",
			DBName:  "ledgerguard",
			SSLMode: "disable",
		},
	}

	// Load from file if provided
	if configPath != "" {
		if err := loadFromFile(configPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables
	applyEnvOverrides(cfg)

	return cfg, nil
}

func loadFromFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse yaml: %w", err)
	}

	return nil
}

func applyEnvOverrides(cfg *Config) {
	// Server
	if v := os.Getenv("SERVER_PORT"); v != "" {
		cfg.Server.Port = v
	}

	// Database
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		cfg.Database.Port = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.DBName = v
	}
	if v := os.Getenv("DB_SSLMODE"); v != "" {
		cfg.Database.SSLMode = v
	}

	// Firebase
	if v := os.Getenv("FIREBASE_CREDENTIALS_FILE"); v != "" {
		cfg.Firebase.CredentialsFile = v
	}
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}
