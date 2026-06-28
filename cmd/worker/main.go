package main

import (
	"encoding/json"
	"food-order-api/internal/models"
	"food-order-api/internal/queue"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	qm, err := queue.NewQueueManager(rabbitURL)
	if err != nil {
		slog.Error("Worker unable to establish connection to Queue", "error", err)
		os.Exit(1)
	}
	defer qm.Close()

	msgs, err := qm.ConsumeEvents()
	if err != nil {
		slog.Error("Worker unable to claim message stream channels", "error", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Worker pipeline processing loop successfully initiated.")
		for msg := range msgs {
			var event models.OrderPlacedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				slog.Error("Unable to safely map data structure parsing payload", "error", err)
				continue
			}

			// Simulated structured processing logging output as required
			slog.Info("✨ KITCHEN DISPLAY TICKER TRIGGERED ✨",
				"order_id", event.OrderID,
				"total_amount_cents", event.TotalCents,
				"received_timestamp", event.Timestamp,
			)
		}
	}()

	<-sigChan
	slog.Info("Shutting down worker process cleanly...")
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}