package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"referee/internal/arbitrage"
	"referee/internal/config"
	"referee/internal/database"
	"referee/internal/exchange"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"referee/internal/model"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("Starting Referee arbitrage simulation bot")

	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	logger.Info("Configuration loaded successfully")

	// Create database connection pool
	connStr := "postgres://" + cfg.Database.User + ":" + cfg.Database.Password + "@" + cfg.Database.Host + ":" + fmt.Sprintf("%d", cfg.Database.Port) + "/" + cfg.Database.DBName
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("Database connection established")

	// Create repository
	repo := &database.PostgresRepository{Pool: pool}

	// Create arbitrage engine
	engine := arbitrage.NewArbitrageEngine(logger, repo, &cfg)
	logger.Info("Arbitrage engine initialized")

	// Create exchange clients
	clients := []exchange.ExchangeClient{
		exchange.NewKrakenClient(logger),
		exchange.NewBinanceClient(logger),
	}
	logger.Info("Exchange clients created", "count", len(clients))

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create the fan-in channel for price ticks
	priceChan := make(chan model.PriceTick, 100)

	// Start the arbitrage engine goroutine
	go func() {
		logger.Info("Starting arbitrage engine")
		for {
			select {
			case <-ctx.Done():
				logger.Info("Arbitrage engine shutting down")
				return
			case tick := <-priceChan:
				engine.ProcessTick(ctx, tick)
			}
		}
	}()

	// Start all exchange clients in goroutines
	for _, client := range clients {
		go func(c exchange.ExchangeClient) {
			logger.Info("Starting exchange client", "exchange", c.GetName())
			if err := c.StartStream(ctx, priceChan, "BTC/EUR"); err != nil {
				logger.Error("Exchange client error", "exchange", c.GetName(), "error", err)
			}
		}(client)
	}

	// Wait for shutdown signal
	logger.Info("Referee is running. Press Ctrl+C to stop.")
	<-sigChan
	logger.Info("Shutdown signal received, initiating graceful shutdown")

	// Cancel context to stop all goroutines
	cancel()

	// Give some time for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Wait for shutdown or timeout
	select {
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout reached, forcing exit")
	default:
		logger.Info("Graceful shutdown completed")
	}
}
