package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any env vars that might interfere
	os.Clearenv()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("expected default port '8080', got '%s'", cfg.Server.Port)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("expected default host 'localhost', got '%s'", cfg.Database.Host)
	}

	if cfg.Database.Port != "5432" {
		t.Errorf("expected default port '5432', got '%s'", cfg.Database.Port)
	}

	if cfg.Database.SSLMode != "disable" {
		t.Errorf("expected default sslmode 'disable', got '%s'", cfg.Database.SSLMode)
	}
}

func TestLoad_FromYAMLFile(t *testing.T) {
	os.Clearenv()

	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
server:
  port: "9090"

database:
  host: "db.example.com"
  port: "5433"
  user: "testuser"
  password: "testpass"
  name: "testdb"
  sslmode: "require"

firebase:
  credentials_file: "/path/to/creds.json"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != "9090" {
		t.Errorf("expected port '9090', got '%s'", cfg.Server.Port)
	}

	if cfg.Database.Host != "db.example.com" {
		t.Errorf("expected host 'db.example.com', got '%s'", cfg.Database.Host)
	}

	if cfg.Database.Port != "5433" {
		t.Errorf("expected port '5433', got '%s'", cfg.Database.Port)
	}

	if cfg.Database.User != "testuser" {
		t.Errorf("expected user 'testuser', got '%s'", cfg.Database.User)
	}

	if cfg.Database.Password != "testpass" {
		t.Errorf("expected password 'testpass', got '%s'", cfg.Database.Password)
	}

	if cfg.Database.SSLMode != "require" {
		t.Errorf("expected sslmode 'require', got '%s'", cfg.Database.SSLMode)
	}

	if cfg.Firebase.CredentialsFile != "/path/to/creds.json" {
		t.Errorf("expected credentials file '/path/to/creds.json', got '%s'", cfg.Firebase.CredentialsFile)
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
server:
  port: "9090"

database:
  host: "db.example.com"
  user: "fileuser"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set env vars to override
	os.Setenv("SERVER_PORT", "7070")
	os.Setenv("DB_USER", "envuser")
	defer os.Clearenv()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Env should override file
	if cfg.Server.Port != "7070" {
		t.Errorf("expected port '7070' (from env), got '%s'", cfg.Server.Port)
	}

	if cfg.Database.User != "envuser" {
		t.Errorf("expected user 'envuser' (from env), got '%s'", cfg.Database.User)
	}

	// File value should be used when no env override
	if cfg.Database.Host != "db.example.com" {
		t.Errorf("expected host 'db.example.com' (from file), got '%s'", cfg.Database.Host)
	}
}

func TestLoad_EnvOnly(t *testing.T) {
	os.Clearenv()
	os.Setenv("SERVER_PORT", "3000")
	os.Setenv("DB_HOST", "envhost")
	os.Setenv("DB_PASSWORD", "envpass")
	os.Setenv("FIREBASE_CREDENTIALS_FILE", "/env/creds.json")
	defer os.Clearenv()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != "3000" {
		t.Errorf("expected port '3000', got '%s'", cfg.Server.Port)
	}

	if cfg.Database.Host != "envhost" {
		t.Errorf("expected host 'envhost', got '%s'", cfg.Database.Host)
	}

	if cfg.Database.Password != "envpass" {
		t.Errorf("expected password 'envpass', got '%s'", cfg.Database.Password)
	}

	if cfg.Firebase.CredentialsFile != "/env/creds.json" {
		t.Errorf("expected credentials file '/env/creds.json', got '%s'", cfg.Firebase.CredentialsFile)
	}
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	expected := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"
	if cfg.DSN() != expected {
		t.Errorf("expected DSN '%s', got '%s'", expected, cfg.DSN())
	}
}
