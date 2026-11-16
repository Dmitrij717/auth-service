package configs

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	Port        string
	Secret      string
	PostgresURL string
}

func LoadConfig() (Config, error) {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	secret, ok := os.LookupEnv("JWT_SECRET")
	if !ok || secret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET not specified")
	}

	user, ok := os.LookupEnv("POSTGRES_USER")
	if !ok || user == "" {
		return Config{}, fmt.Errorf("POSTGRES_USER not specified")
	}

	password, ok := os.LookupEnv("POSTGRES_PASSWORD")
	if !ok || password == "" {
		return Config{}, fmt.Errorf("POSTGRES_PASSWORD not specified")
	}

	postgresHost, ok := os.LookupEnv("POSTGRES_HOST")
	if !ok || postgresHost == "" {
		postgresHost = "localhost"
	}

	postgresPort, ok := os.LookupEnv("POSTGRES_PORT")
	if !ok || postgresPort == "" {
		return Config{}, fmt.Errorf("POSTGRES_PORT not specified")
	}

	db, ok := os.LookupEnv("POSTGRES_DB")
	if !ok || db == "" {
		return Config{}, fmt.Errorf("POSTGRES_DB not specified")
	}

	postgresURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     postgresHost + ":" + postgresPort,
		RawQuery: "sslmode=disable",
		Path:     "/" + db,
	}

	return Config{
		Port:        port,
		Secret:      secret,
		PostgresURL: postgresURL.String(),
	}, nil
}
