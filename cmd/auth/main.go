package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/Dmitrij717/auth-service/configs"
	"github.com/Dmitrij717/auth-service/internal/repository"
	"github.com/Dmitrij717/auth-service/internal/service"
	"github.com/Dmitrij717/auth-service/internal/token"
	http_handlers "github.com/Dmitrij717/auth-service/internal/transport/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("warning: .env file not loaded: %v", err)
	}

	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("load config - %v", err)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, config.PostgresURL)
	if err != nil {
		log.Fatalf("failed connection to database - %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping database - %v", err)
	}

	repo := repository.NewPostgresRepo(pool)
	jwtManager := token.NewJWTManager(config.Secret)
	service := service.NewUserService(repo, jwtManager)
	loginHandler := http_handlers.NewLoginHandler(service)
	verifyHandler := http_handlers.NewVerifyHandler(service)
	registerHandler := http_handlers.NewRegisterHandler(service)

	http.HandleFunc("/login", loginHandler.Handle)
	http.HandleFunc("/verify", verifyHandler.Handle)
	http.HandleFunc("/register", registerHandler.Handle)

	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatalf("server error - %v", err)
	}
}
