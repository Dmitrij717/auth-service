package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Dmitrij717/auth-service/configs"
	"github.com/Dmitrij717/auth-service/internal/metrics"
	"github.com/Dmitrij717/auth-service/internal/middleware"
	"github.com/Dmitrij717/auth-service/internal/repository"
	"github.com/Dmitrij717/auth-service/internal/service"
	"github.com/Dmitrij717/auth-service/internal/token"
	grpc_server "github.com/Dmitrij717/auth-service/internal/transport/grpc"
	"github.com/Dmitrij717/auth-service/internal/transport/grpc/gen/auth"
	http_handlers "github.com/Dmitrij717/auth-service/internal/transport/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func main() {

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("warning: .env file not loaded: %v", err)
	}

	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("load config - %v", err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

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

	metrics, err := metrics.NewMetrics()
	if err != nil {
		log.Fatalf("metrics initialization: %v", err)
	}

	metricServer := http.Server{
		Addr:    ":2112",
		Handler: metrics.Handler(),
	}

	metricsLis, err := net.Listen("tcp", metricServer.Addr)
	if err != nil {
		log.Fatalf("listening metrics port: %v", err)
	}

	go func() {
		if err := metricServer.Serve(metricsLis); err != nil {
			log.Printf("metrics server: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/register", registerHandler.Handle)
	mux.HandleFunc("/login", loginHandler.Handle)
	mux.HandleFunc("/verify", verifyHandler.Handle)

	handler := middleware.HTTPMetricMiddleware(metrics, mux)

	httpServer := http.Server{
		Addr:    ":" + config.Port,
		Handler: handler,
	}

	grpcHandler := grpc_server.NewServer(service)
	grpcServer := grpc.NewServer()
	auth.RegisterAuthServiceServer(grpcServer, grpcHandler)

	grpcLis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen - %v", err)
	}
	defer func() {
		if err := grpcLis.Close(); err != nil {
			log.Printf("closing gRPC-Listener: %v", err)
		}
	}()

	httpLis, err := net.Listen("tcp", httpServer.Addr)
	if err != nil {
		log.Fatalf("failed to listen - %v", err)
	}
	defer func() {
		if err := httpLis.Close(); err != nil {
			log.Printf("closing HTTP-listener: %v", err)
		}
	}()

	go func() {
		if err := httpServer.Serve(httpLis); err != nil {
			log.Printf("HTTP server failed: %v", err)
		}
	}()

	go func() {
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Printf("gRPC server failed: %v", err)
		}

	}()

	g, errGroupCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := httpServer.Serve(httpLis)
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	})

	g.Go(func() error {
		err := grpcServer.Serve(grpcLis)
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return err
	})

	g.Go(func() error {
		<-errGroupCtx.Done()

		shutdownCtx, cancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			10*time.Second,
		)
		defer cancel()

		var wg sync.WaitGroup

		wg.Add(3)

		go func() {
			defer wg.Done()
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				log.Printf("Shutdown HTTP Server: %v", err)

				if closeErr := httpServer.Close(); closeErr != nil {
					log.Printf("Close HTTP Server: %v", closeErr)
				}
			}
		}()

		go func() {
			defer wg.Done()
			if err := metricServer.Shutdown(shutdownCtx); err != nil {
				log.Printf("Shutdown Metrics Server: %v", err)
			}

			if err := metricServer.Close(); err != nil {
				log.Printf("Close Metrics Server: %v", err)
			}
		}()

		grpcDone := make(chan struct{})
		timer := time.NewTimer(10 * time.Second)
		defer timer.Stop()

		go func() {
			defer wg.Done()
			grpcServer.GracefulStop()
			close(grpcDone)
		}()

		select {
		case <-timer.C:
			grpcServer.Stop()
		case <-grpcDone:
			log.Printf("gRPC Server end gracefully")
		}

		<-grpcDone

		wg.Wait()

		return nil
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

}
