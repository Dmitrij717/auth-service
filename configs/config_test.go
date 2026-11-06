package configs

import (
	"os"
	"testing"
)

func TestLoadConfig_WithPort(t *testing.T) {
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
