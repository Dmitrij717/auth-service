package configs

import "os"

type Config struct {
	Port string
}

func LoadConfig() (Config, error) {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	return Config{Port: port}, nil
}
