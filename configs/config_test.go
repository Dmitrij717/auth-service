package configs

import (
	"os"
	"testing"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()

	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("POSTGRES_USER", "test-user")
	t.Setenv("POSTGRES_PASSWORD", "test-password")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_DB", "test-db")
}

func TestLoadConfig_WithPort(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("PORT", "9090")

	conf, err := LoadConfig()
	if err != nil {
		// прекращение работы пока искусственное
		t.Fatalf("LoadConfig() returned unexpected error: %v", err)
	}

	if conf.Port != "9090" {
		t.Fatalf("expected port 9090, got %v", conf.Port)
	}

}

func TestLoadConfig_DefaultPort(t *testing.T) {
	setRequiredEnv(t)
	val, ok := os.LookupEnv("PORT")
	t.Cleanup(func() {
		if ok {
			if err := os.Setenv("PORT", val); err != nil {
				t.Errorf("os.Setenv() error - %v", err)
			}
		} else {
			if err := os.Unsetenv("PORT"); err != nil {
				t.Errorf("os.Unsetenv() error - %v", err)
			}
		}
	})

	err := os.Unsetenv("PORT")
	if err != nil {
		t.Fatalf("os.Unsetenv() error - %v", err)
	}

	conf, err := LoadConfig()
	if err != nil {
		// прекращение работы пока искусственное
		t.Fatalf("LoadConfig() returned unexpected error: %v", err)
	}

	if conf.Port != "8080" {
		t.Fatalf("expected port 8080, got %v", conf.Port)
	}
}
