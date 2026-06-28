package main

import (
	"context"
	"food-order-api/internal/handlers"
	"food-order-api/internal/queue"
	"food-order-api/internal/repository"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/food_db?sslmode=disable")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	var pool *pgxpool.Pool
	var err error
	for i := 0; i < 5; i++ { // Quick retry buffer for slow DB containers
		pool, err = pgxpool.New(ctx, dbURL)
		if err == nil && pool.Ping(ctx) == nil {
			break
		}
		slog.Warn("Waiting for database connection...")
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		slog.Error("Failed to open connection to Database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	qm, err := queue.NewQueueManager(rabbitURL)
	if err != nil {
		slog.Error("Failed to open connection to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer qm.Close()

	repo := repository.NewRepository(pool)
	h := handlers.NewHandler(repo, qm)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/menu", h.GetMenu)
	r.Get("/menu/items/{id}", h.GetMenuItem)
	r.Patch("/menu/items/{id}", h.UpdateItemAvailability)

	r.Post("/orders", h.PlaceOrder)
	r.Get("/orders/{id}", h.GetOrder)
	r.Patch("/orders/{id}/status", h.UpdateOrderStatus)

	slog.Info("HTTP service running successfully on port :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		slog.Error("Server crushed unexpectedly", "error", err)
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}