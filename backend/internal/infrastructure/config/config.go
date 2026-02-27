package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Firebase   FirebaseConfig   `yaml:"firebase"`
	Shopify    ShopifyConfig    `yaml:"shopify"`
	Encryption EncryptionConfig `yaml:"encryption"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type DatabaseConfig struct {
	Host           string `yaml:"host"`
	Port           string `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	DBName         string `yaml:"name"`
	SSLMode        string `yaml:"sslmode"`
	MigrationsPath string `yaml:"migrations_path"`
}

type FirebaseConfig struct {
	CredentialsFile string `yaml:"credentials_file"`
}

type ShopifyConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
	Scopes       string `yaml:"scopes"`
}

type EncryptionConfig struct {
	MasterKey string `yaml:"master_key"`
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
			Host:           "localhost",
			Port:           "5432",
			User:           "postgres",
			DBName:         "ledgerguard",
			SSLMode:        "disable",
			MigrationsPath: "migrations",
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
	if v := os.Getenv("DB_MIGRATIONS_PATH"); v != "" {
		cfg.Database.MigrationsPath = v
	}

	// Firebase
	if v := os.Getenv("FIREBASE_CREDENTIALS_FILE"); v != "" {
		cfg.Firebase.CredentialsFile = v
	}

	// Shopify
	if v := os.Getenv("SHOPIFY_CLIENT_ID"); v != "" {
		cfg.Shopify.ClientID = v
	}
	if v := os.Getenv("SHOPIFY_CLIENT_SECRET"); v != "" {
		cfg.Shopify.ClientSecret = v
	}
	if v := os.Getenv("SHOPIFY_REDIRECT_URI"); v != "" {
		cfg.Shopify.RedirectURI = v
	}
	if v := os.Getenv("SHOPIFY_SCOPES"); v != "" {
		cfg.Shopify.Scopes = v
	}

	// Encryption
	if v := os.Getenv("ENCRYPTION_MASTER_KEY"); v != "" {
		cfg.Encryption.MasterKey = v
	}
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}
