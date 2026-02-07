package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ecos/ingestion/internal/api"
	"github.com/ecos/ingestion/internal/config"
	"github.com/ecos/ingestion/internal/events"
	"github.com/ecos/ingestion/internal/store"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	// DynamoDB client.
	dynamoClient, err := store.NewDynamoDBClient(ctx, cfg.AWSRegion, cfg.DynamoDBEndpoint)
	if err != nil {
		slog.Error("failed to create DynamoDB client", "error", err)
		os.Exit(1)
	}

	// SNS client.
	snsClient, err := events.NewSNSClient(ctx, cfg.AWSRegion, cfg.SNSEndpoint)
	if err != nil {
		slog.Error("failed to create SNS client", "error", err)
		os.Exit(1)
	}

	// Stores.
	readingsStore := store.NewReadingsStore(dynamoClient, cfg.ReadingsTableName, cfg.ReadingTypeIndexName)
	stationsStore := store.NewStationsStore(dynamoClient, cfg.StationsTableName)

	// Publisher.
	publisher := events.NewPublisher(snsClient, cfg.DataIngestionTopicARN, cfg.AlertEventsTopicARN)

	// Router.
	router := api.NewRouter(readingsStore, stationsStore, publisher)

	// HTTP server.
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("starting server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-shutdown
	slog.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
