package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexandrepossebom/rinha-backend-2024/handlers"
	"github.com/alexandrepossebom/rinha-backend-2024/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var err error

	port := os.Getenv("PORT")
	dsn := os.Getenv("DATABASE_URL")
	if port == "" || dsn == "" {
		log.Fatal("PORT and DATABASE_URL must be set")
	}

	dbConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal(err)
	}

	dbConfig.MaxConns = 5
	dbConfig.MinConns = 2
	connPool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer connPool.Close()

	if err := connPool.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to the database!!")

	handler := handlers.NewHandler(repo.NewRepository(connPool))
	mux := http.NewServeMux()

	mux.HandleFunc("POST /clientes/{id}/transacoes", handler.NewTransacaoHandler)
	mux.HandleFunc("GET /clientes/{id}/extrato", handler.NewExtratoHandler)

	srv := http.Server{Addr: fmt.Sprintf(":%s", port), Handler: mux, ReadHeaderTimeout: 1 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen err: %v", err)
		}
	}()

	log.Println("Server started")

	<-ctx.Done()
	if err := srv.Shutdown(context.TODO()); err != nil {
		log.Println(err)
	}
	log.Println("Server stopped")
}
